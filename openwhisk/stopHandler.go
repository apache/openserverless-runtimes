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
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type stopRequest struct {
	ActionCodeHash  string `json:"actionCodeHash,omitempty"`
	ProxiedActionID string `json:"proxiedActionID,omitempty"`
}

var setupActionPath = "tmp"

func (ap *ActionProxy) stopHandler(w http.ResponseWriter, r *http.Request) {
	if ap.proxyMode != ProxyModeServer {
		sendError(w, http.StatusUnprocessableEntity, "Stop is only supported in server mode")
		return
	}

	// parse the request
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("Error reading request body: %v", err))
		return
	}

	var stopRequest stopRequest
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&stopRequest)
	if err != nil {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("Error decoding run body: %v", err))
		return
	}

	innerAPValue, ok := ap.serverProxyData.actions[stopRequest.ActionCodeHash]
	if !ok {
		Debug("Action hash '%s' not found in server proxy data", stopRequest.ActionCodeHash)
		sendError(w, http.StatusNotFound, "Action to be removed in remote runtime not found. Check logs for details.")
		return
	}

	connectedIDFound := false
	for i, connectedID := range innerAPValue.connectedActionIDs {
		if connectedID == stopRequest.ProxiedActionID {
			// remove id from the array
			innerAPValue.connectedActionIDs = removeID(innerAPValue.connectedActionIDs, i)
			connectedIDFound = true
			break
		}
	}
	if !connectedIDFound {
		Debug("Action ID '%s' not found in server proxy data", stopRequest.ProxiedActionID)
		sendError(w, http.StatusNotFound, "Action to be removed in remote runtime not found. Check logs for details.")
		return
	}

	Debug("Removed action ID. Length of connectedActionIDs: %d", len(innerAPValue.connectedActionIDs))

	if isSetupActionRunning(stopRequest.ActionCodeHash) {
		// start timer
		// TODO
		sendOK(w)
		return
	}
	if len(innerAPValue.connectedActionIDs) == 0 {
		Debug("Action hash '%s' executor stopped", stopRequest.ActionCodeHash)
		close(innerAPValue.runRequestQueue)
		cleanUpAP(innerAPValue.remoteProxy)
		delete(ap.serverProxyData.actions, stopRequest.ActionCodeHash)
	}

	sendOK(w)
}

func cleanUpAP(ap *ActionProxy) {
	ap.theExecutor.Stop()
	if err := os.RemoveAll(filepath.Join(ap.baseDir, strconv.Itoa(ap.currentDir))); err != nil {
		Debug("Error removing action directory: %v", err)
	}
}

func removeID(idArray []string, index int) []string {
	ret := make([]string, 0)
	ret = append(ret, idArray[:index]...)
	return append(ret, idArray[index+1:]...)
}

func isSetupActionRunning(actionCodeHash string) bool {
	// A setup action is running if
	// - the file "/tmp/{hash}" exists
	// - the file "/tmp/{hash}_done" does not exist

	path := filepath.Join(setupActionPath, actionCodeHash)
	_, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return false
	}

	setupDoneFile := path + "_done"
	_, err = os.Stat(setupDoneFile)
	return errors.Is(err, fs.ErrNotExist)
}
