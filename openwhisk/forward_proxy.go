/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package openwhisk

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

const OW_CODE_HASH = "__OW_CODE_HASH"

func (ap *ActionProxy) ForwardRunRequest(w http.ResponseWriter, r *http.Request) {
	if ap.clientProxyData == nil {
		sendError(w, http.StatusInternalServerError, "Send init first")
		return
	}
	var runRequest runRequest
	err := json.NewDecoder(r.Body).Decode(&runRequest)
	if err != nil {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("Error decoding run body while forwarding request: %v", err))
		return
	}

	newBody := runRequest
	newBody.ProxiedActionID = ap.clientProxyData.ProxyActionID

	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(newBody)
	if err != nil {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("Error encoding updated init body: %v", err))
		return
	}

	bodyLen := buf.Len()
	r.Body = io.NopCloser(bytes.NewBuffer(buf.Bytes()))

	director := func(req *http.Request) {
		req.Header = r.Header.Clone()

		// Reset content length with the new body
		req.Header.Set("Content-Length", strconv.Itoa(bodyLen))
		req.ContentLength = int64(bodyLen)

		req.URL.Scheme = ap.clientProxyData.ProxyURL.Scheme
		req.URL.Host = ap.clientProxyData.ProxyURL.Host
		req.Host = ap.clientProxyData.ProxyURL.Host
	}

	proxy := &httputil.ReverseProxy{Director: director}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		Debug("Error forwarding run request: %v", err)
		sendError(w, http.StatusBadGateway, "Error forwarding run request. Check logs for details.")
	}

	proxy.ModifyResponse = func(response *http.Response) error {
		if response.StatusCode == http.StatusOK {
			// Decode the response
			var remoteReponse RemoteRunResponse
			err := json.NewDecoder(response.Body).Decode(&remoteReponse)
			if err != nil {
				Debug("Error decoding remote response: %v", err)
				return err
			}

			// Write the logs to the client logs
			if _, err := ap.outFile.WriteString(remoteReponse.Out); err != nil {
				Debug("Error writing remote response out to client: %v", err)
			}
			if _, err := ap.errFile.WriteString(remoteReponse.Err); err != nil {
				Debug("Error writing remote response err to client: %v", err)
			}

			// Keep the response body only
			response.Body = io.NopCloser(bytes.NewReader(remoteReponse.Response))
		}

		return nil
	}

	Debug("Forwarding run request with id %s to %s", newBody.ProxiedActionID, ap.clientProxyData.ProxyURL.String())
	proxy.ServeHTTP(w, r)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

func (ap *ActionProxy) ForwardInitRequest(w http.ResponseWriter, r *http.Request) {
	var initRequest initRequest
	err := json.NewDecoder(r.Body).Decode(&initRequest)
	if err != nil {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("Error decoding init body while forwarding request: %v", err))
		return
	}

	Debug("Decoded init request: len: %d - main: %s", r.ContentLength, initRequest.Value.Main)
	proxyData, err := parseMainFlag(initRequest.Value.Main)
	if err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// set the proxy data
	ap.clientProxyData = proxyData
	ap.clientProxyData.ProxyActionID = uuid.New().String()

	newBody := initRequest
	newBody.Value.Main = ap.clientProxyData.MainFunc
	newBody.ProxiedActionID = ap.clientProxyData.ProxyActionID
	codeHash := calculateCodeHash(initRequest.Value.Code)
	if newBody.Value.Env == nil {
		newBody.Value.Env = make(map[string]interface{})
	}
	newBody.Value.Env[OW_CODE_HASH] = codeHash

	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(newBody)
	if err != nil {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("Error encoding updated init body: %v", err))
		return
	}

	bodyLen := buf.Len()
	r.Body = io.NopCloser(bytes.NewBuffer(buf.Bytes()))

	director := func(req *http.Request) {
		req.Header = r.Header.Clone()

		// Reset content length with the new body
		req.Header.Set("Content-Length", strconv.Itoa(bodyLen))
		req.ContentLength = int64(bodyLen)

		req.URL.Scheme = ap.clientProxyData.ProxyURL.Scheme
		req.URL.Host = ap.clientProxyData.ProxyURL.Host
		req.Host = ap.clientProxyData.ProxyURL.Host
	}

	proxy := &httputil.ReverseProxy{Director: director}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		Debug("Error forwarding init request: %v", err)
		sendError(w, http.StatusBadGateway, "Error forwarding init request. Check logs for details.")
	}

	Debug("Forwarding init request to %s", ap.clientProxyData.ProxyURL.String())
	proxy.ServeHTTP(w, r)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

func parseMainFlag(mainAtProxy string) (*ClientProxyData, error) {
	proxyData := ClientProxyData{}
	splitedMainAtProxy := strings.Split(mainAtProxy, "@")

	var extractedURL string

	if len(splitedMainAtProxy) == 2 {
		proxyData.MainFunc = splitedMainAtProxy[0]
		extractedURL = splitedMainAtProxy[1]
	} else if len(splitedMainAtProxy) == 1 {
		extractedURL = splitedMainAtProxy[0]
	} else {
		return nil, fmt.Errorf("invalid value for --main flag. Must be in the form of <main>@<proxy> or @<proxy>")
	}

	parsedUrl, err := parseMainURL(extractedURL)
	if err != nil {
		return nil, err
	}

	proxyData.ProxyURL = *parsedUrl

	Debug("Parsed main flag. Main: %s, Proxy: %s", proxyData.MainFunc, proxyData.ProxyURL.String())
	return &proxyData, nil
}

func parseMainURL(input string) (*url.URL, error) {
	if input == "" {
		return nil, fmt.Errorf("empty URL")
	}
	// Check if the input has a scheme, otherwise "https"
	if !strings.Contains(input, "://") {
		input = "https://" + input
	}

	// Parse the input URL
	parsedURL, err := url.Parse(input)
	if err != nil {
		return nil, err
	}

	return parsedURL, nil
}

func calculateCodeHash(code string) string {
	hash := md5.Sum([]byte(code))
	return hex.EncodeToString(hash[:])
}
