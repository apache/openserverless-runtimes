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

type remoteRunChanPayload struct {
	runRequest *runRequest
	respChan   chan *ServerRunResponseChanPayload
}
type ServerRunResponseChanPayload struct {
	runResp *RemoteRunResponse
	status  int
	err     error
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
	if ap.proxyMode == ProxyModeNone {
		ap.doRun(w, r)
		return
	}

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

		if runRequest.ActionCodeHash == "" {
			sendError(w, http.StatusBadRequest, "Action code hash not provided from client")
			return
		}
		innerActionProxy, ok := ap.serverProxyData.actions[runRequest.ActionCodeHash]
		if !ok {
			Debug("Action hash %s not found in server proxy data", runRequest.ActionCodeHash)
			sendError(w, http.StatusNotFound, "Action not found in remote runtime. Check logs for details.")
			return
		}

		// Enqueue the request to be processed by the inner proxy one at a time
		responseChan := make(chan *ServerRunResponseChanPayload)

		innerActionProxy.runRequestQueue <- &remoteRunChanPayload{runRequest: &runRequest, respChan: responseChan}

		res := <-responseChan
		if res.err != nil {
			sendError(w, res.status, res.err.Error())
			return
		}

		// write response
		// turn response struct into json
		responsePayload, err := json.Marshal(res.runResp)
		if err != nil {
			sendError(w, http.StatusInternalServerError, fmt.Sprintf("Error marshalling response: %v", err))
			return
		}

		// write response
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(responsePayload)))
		numBytesWritten, err := w.Write(responsePayload)

		// handle writing errors
		if err != nil {
			Debug("(remote) Error writing response: %v", err)
			sendError(w, http.StatusInternalServerError, fmt.Sprintf("Error writing response: %v", err))
			return
		}
		if numBytesWritten < len(responsePayload) {
			Debug("(remote) Only wrote %d of %d bytes to response", numBytesWritten, len(responsePayload))
			sendError(w, http.StatusInternalServerError, fmt.Sprintf("Only wrote %d of %d bytes to response", numBytesWritten, len(responsePayload)))
			return
		}

		// flush output if possible
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		close(responseChan)
	}
}

func startListenToRunRequests(ap *ActionProxy, runRequestQueue chan *remoteRunChanPayload) {
	for runReq := range runRequestQueue {
		remoteResponse, status, err := ap.doServerModeRun(runReq.runRequest)
		runReq.respChan <- &ServerRunResponseChanPayload{runResp: &remoteResponse, status: status, err: err}
	}
}

func (ap *ActionProxy) doServerModeRun(bodyRequest *runRequest) (RemoteRunResponse, int, error) {
	Debug("Executing run request in server mode")
	body, status, err := prepareRemoteRunBody(ap, bodyRequest)
	if err != nil {
		return RemoteRunResponse{}, status, err
	}

	// execute the action
	response, err := ap.theExecutor.Interact(body)

	// check for early termination
	if err != nil {
		Debug("WARNING! Command exited")
		ap.theExecutor = nil
		return RemoteRunResponse{}, http.StatusBadRequest, fmt.Errorf("command exited")
	}
	DebugLimit("received (remote): ", response, 120)

	// check if the answer is an object map or array
	if ok := isJsonObjOrArray(response); !ok {
		return RemoteRunResponse{}, http.StatusBadGateway, fmt.Errorf("the action did not return a dictionary or array")
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

	return remoteResponse, 0, nil
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

func prepareRemoteRunBody(ap *ActionProxy, bodyRequest *runRequest) ([]byte, int, error) {
	var bodyBuf bytes.Buffer
	err := json.NewEncoder(&bodyBuf).Encode(bodyRequest)
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("error encoding proxied run body: %v", err)
	}
	body := bytes.Replace(bodyBuf.Bytes(), []byte("\n"), []byte(""), -1)

	// check if you have an action
	if ap.theExecutor == nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("no action defined yet")
	}

	// check if the process exited
	if ap.theExecutor.Exited() {
		return nil, http.StatusInternalServerError, fmt.Errorf("command exited")
	}

	return body, 0, nil
}
