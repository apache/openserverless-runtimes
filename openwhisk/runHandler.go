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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type runRequest struct {
	ActionCodeHash string                 `json:"actionCodeHash,omitempty"`
	Value          map[string]interface{} `json:"value,omitempty"`
}

// ErrResponse is the response when there are errors
type ErrResponse struct {
	Error string `json:"error"`
}

type RemoteRunResponse struct {
	Response json.RawMessage `json:"response"`
	Out      string          `json:"out"`
	Err      string          `json:"err"`
}

func sendError(w http.ResponseWriter, code int, cause string) {
	errResponse := ErrResponse{Error: cause}
	b, err := json.Marshal(errResponse)
	if err != nil {
		b = []byte("error marshalling error response")
		Debug(err.Error())
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(b)
	w.Write([]byte("\n"))
}

func (ap *ActionProxy) runHandler(w http.ResponseWriter, r *http.Request) {
	if ap.proxyMode == ProxyModeClient {
		ap.ForwardRunRequest(w, r)
		return
	}

	if ap.proxyMode == ProxyModeServer {
		// parse the request
		body, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			sendError(w, http.StatusBadRequest, fmt.Sprintf("Error reading request body: %v", err))
			return
		}

		var runRequest runRequest
		err = json.NewDecoder(bytes.NewReader(body)).Decode(&runRequest)
		if err != nil {
			sendError(w, http.StatusBadRequest, fmt.Sprintf("Error decoding run body: %v", err))
			return
		}

		actionHash := runRequest.ActionCodeHash
		innerActionProxy, ok := ap.serverProxyData.actions[actionHash]
		if !ok {
			Debug("Action hash %s not found in server proxy data", actionHash)
			sendError(w, http.StatusNotFound, "Action not found in remote runtime. Check logs for details.")
			return
		}

		innerActionProxy.remoteProxy.doServerModeRun(w, &runRequest)
		return
	}

	ap.doRun(w, r)
}

func (ap *ActionProxy) doServerModeRun(w http.ResponseWriter, bodyRequest *runRequest) {
	body, ok := prepareRemoteRunBody(ap, w, bodyRequest)
	if !ok {
		return
	}

	// execute the action
	response, err := ap.theExecutor.Interact(body)

	// check for early termination
	if err != nil {
		Debug("WARNING! Command exited")
		ap.theExecutor = nil
		sendError(w, http.StatusBadRequest, "command exited")
		return
	}
	DebugLimit("received (remote):", response, 120)

	// check if the answer is an object map or array
	if ok := isJsonObjOrArray(response); !ok {
		sendError(w, http.StatusBadGateway, "The action did not return a dictionary or array.")
		return
	}

	// Get the stdout and stderr from the executor
	outStr, err := os.ReadFile(ap.outFile.Name())
	if err != nil {
		outStr = []byte(fmt.Sprintf("Error reading stdout: %v", err))
	}

	errStr, err := os.ReadFile(ap.errFile.Name())
	if err != nil {
		errStr = []byte(fmt.Sprintf("Error reading stderr: %v", err))
	}

	// create the response struct
	remoteResponse := RemoteRunResponse{
		Response: response,
		Out:      string(outStr),
		Err:      string(errStr),
	}

	// turn response struct into json
	responsePayload, err := json.Marshal(remoteResponse)
	if err != nil {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("Error marshalling response: %v", err))
		return
	}

	// write response
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(responsePayload)))
	numBytesWritten, err := w.Write(responsePayload)

	// flush output if possible
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	// handle writing errors
	if err != nil {
		Debug("(remote) Error writing response: %v", err)
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("Error writing response: %v", err))
		return
	}
	if numBytesWritten != len(response) {
		Debug("(remote) Only wrote %d of %d bytes to response", numBytesWritten, len(response))
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("Only wrote %d of %d bytes to response", numBytesWritten, len(response)))
		return
	}
}

func (ap *ActionProxy) doRun(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("Error reading request body: %v", err))
		return
	}
	Debug("done reading %d bytes", len(body))

	body = bytes.Replace(body, []byte("\n"), []byte(""), -1)

	// check if you have an action
	if ap.theExecutor == nil {
		sendError(w, http.StatusInternalServerError, "no action defined yet")
		return
	}

	// check if the process exited
	if ap.theExecutor.Exited() {
		sendError(w, http.StatusInternalServerError, "command exited")
		return
	}

	// execute the action
	response, err := ap.theExecutor.Interact(body)

	// check for early termination
	if err != nil {
		Debug("WARNING! Command exited")
		ap.theExecutor = nil
		sendError(w, http.StatusBadRequest, "command exited")
		return
	}
	DebugLimit("received:", response, 120)

	// check if the answer is an object map or array
	if ok := isJsonObjOrArray(response); !ok {
		sendError(w, http.StatusBadGateway, "The action did not return a dictionary or array.")
		return
	}

	// write response
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(response)))
	numBytesWritten, err := w.Write(response)

	// flush output if possible
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	// handle writing errors
	if err != nil {
		Debug("Error writing response: %v", err)
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("Error writing response: %v", err))
		return
	}
	if numBytesWritten != len(response) {
		Debug("Only wrote %d of %d bytes to response", numBytesWritten, len(response))
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("Only wrote %d of %d bytes to response", numBytesWritten, len(response)))
		return
	}
}

func isJsonObjOrArray(response []byte) bool {
	var objmap map[string]*json.RawMessage
	var objarray []interface{}
	err := json.Unmarshal(response, &objmap)
	if err != nil {
		err = json.Unmarshal(response, &objarray)
		if err != nil {
			return false
		}
	}
	return true
}

func prepareRemoteRunBody(ap *ActionProxy, w http.ResponseWriter, bodyRequest *runRequest) ([]byte, bool) {
	var bodyBuf bytes.Buffer
	err := json.NewEncoder(&bodyBuf).Encode(bodyRequest)
	if err != nil {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("Error encoding proxied run body: %v", err))
		return nil, false
	}
	body := bytes.Replace(bodyBuf.Bytes(), []byte("\n"), []byte(""), -1)

	// check if you have an action
	if ap.theExecutor == nil {
		sendError(w, http.StatusInternalServerError, "no action defined yet")
		return nil, false
	}

	// check if the process exited
	if ap.theExecutor.Exited() {
		sendError(w, http.StatusInternalServerError, "command exited")
		return nil, false
	}

	return body, true
}
