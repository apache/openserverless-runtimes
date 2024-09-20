package openwhisk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
)

func (ap *ActionProxy) ForwardRunRequest(w http.ResponseWriter, r *http.Request) {
	if ap.proxyData == nil {
		sendError(w, http.StatusInternalServerError, "Send init first")
		return
	}
	Debug("Forwarding run request to %s", ap.proxyData.ProxyURL.String())
	proxy := httputil.NewSingleHostReverseProxy(&ap.proxyData.ProxyURL)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		sendError(w, http.StatusBadGateway, "Error proxying request: "+err.Error())
	}
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
	ap.proxyData = proxyData

	newBody := initRequest
	newBody.Value.Main = ap.proxyData.MainFunc

	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(newBody)
	if err != nil {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("Error encoding updated init body: %v", err))
		return
	}

	bodyLen := buf.Len()
	Debug("Encoded updated init request: len: %d - main: %s", bodyLen, newBody.Value.Main)

	r.Body = io.NopCloser(bytes.NewBuffer(buf.Bytes()))

	director := func(req *http.Request) {
		req.Header = r.Header.Clone()

		req.Header.Set("Content-Length", strconv.Itoa(bodyLen))
		req.ContentLength = int64(bodyLen)

		req.URL.Scheme = ap.proxyData.ProxyURL.Scheme
		req.URL.Host = ap.proxyData.ProxyURL.Host
	}

	proxy := &httputil.ReverseProxy{Director: director}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		Debug("Error from forward proxy error handler: %v", err)
		sendError(w, http.StatusBadGateway, err.Error())
	}

	Debug("Forwarding init request to %s", ap.proxyData.ProxyURL.String())
	proxy.ServeHTTP(w, r)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

func parseMainFlag(mainAtProxy string) (*ProxyData, error) {
	proxyData := ProxyData{}
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
