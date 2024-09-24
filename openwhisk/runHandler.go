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
)

type runRequest struct {
	ProxiedActionID string                 `json:"proxiedActionID,omitempty"`
	Value           map[string]interface{} `json:"value,omitempty"`
}

// ErrResponse is the response when there are errors
type ErrResponse struct {
	Error string `json:"error"`
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

		actionID := runRequest.ProxiedActionID

		innerActionProxy, ok := ap.ServerProxyData.actions[actionID]
		if !ok {
			Debug("Action %s not found in server proxy data", actionID)
			sendError(w, http.StatusNotFound, "Action not found in remote runtime. Check logs for details.")
		}

		innerActionProxy.doProxiedRun(w, &runRequest)
		return
	}

	ap.doRun(w, r)
}
func (ap *ActionProxy) doProxiedRun(w http.ResponseWriter, bodyRequest *runRequest) {
	var bodyBuf bytes.Buffer
	err := json.NewEncoder(&bodyBuf).Encode(bodyRequest)
	if err != nil {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("Error encoding proxied run body: %v", err))
		return
	}
	body := bytes.Replace(bodyBuf.Bytes(), []byte("\n"), []byte(""), -1)

	ap.executeAction(w, body)
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

	ap.executeAction(w, body)
}

func (ap *ActionProxy) executeAction(w http.ResponseWriter, body []byte) {
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
	var objmap map[string]*json.RawMessage
	var objarray []interface{}
	err = json.Unmarshal(response, &objmap)
	if err != nil {
		err = json.Unmarshal(response, &objarray)
		if err != nil {
			sendError(w, http.StatusBadGateway, "The action did not return a dictionary or array.")
			return
		}
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
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("Error writing response: %v", err))
		return
	}
	if numBytesWritten != len(response) {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("Only wrote %d of %d bytes to response", numBytesWritten, len(response)))
		return
	}
}
