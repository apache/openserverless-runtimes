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
	"time"
)

type stopRequest struct {
	ActionCodeHash  string `json:"actionCodeHash,omitempty"`
	ProxiedActionID string `json:"proxiedActionID,omitempty"`
}

var setupActionPath = "/tmp"

func (ap *ActionProxy) stopHandler(w http.ResponseWriter, r *http.Request) {
	if ap.proxyMode != ProxyModeServer {
		sendError(w, http.StatusUnprocessableEntity, "Stop is only supported in server mode")
		return
	}

	if ap.serverProxyData == nil {
		Debug("Server proxy data not initialized... a restart might have happened!")
		sendError(w, http.StatusInternalServerError, "Server proxy data not initialized")
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

	if len(innerAPValue.connectedActionIDs) == 0 {
		if isSetupActionRunning(stopRequest.ActionCodeHash) {
			go ap.timedDelete(stopRequest.ActionCodeHash)
		} else {
			stopAndDelete(ap, innerAPValue, stopRequest.ActionCodeHash)
		}
	}

	sendOK(w)
}

func stopAndDelete(ap *ActionProxy, innerAPValue *RemoteAPValue, actionCodeHash string) {
	Debug("Action hash '%s' executor stopped", actionCodeHash)
	close(innerAPValue.runRequestQueue)
	cleanUpAP(innerAPValue.remoteProxy)
	delete(ap.serverProxyData.actions, actionCodeHash)
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

	path, err := filepath.Abs(filepath.Join(setupActionPath, actionCodeHash))
	if err != nil {
		Debug("Error getting 'setup check file' absolute path: %v", err)
		return false
	}
	_, err = os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return false
	}

	setupDoneFile := path + "_done"
	_, err = os.Stat(setupDoneFile)
	return errors.Is(err, fs.ErrNotExist)
}

var timeToDeletion = 10 * time.Minute

// timedDelete waits for a certain amount of time before deleting the action.
// The deletion is only done if no new actions have joined.
// The timer duration can be set using the OW_DELETE_DURATION environment variable.
//
// A duration string is a possibly signed sequence of decimal numbers,
// each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m".
// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
func (serverAp *ActionProxy) timedDelete(actionCodeHash string) {
	innerAPValue, ok := serverAp.serverProxyData.actions[actionCodeHash]
	if !ok {
		return
	}

	if len(innerAPValue.connectedActionIDs) > 0 {
		return
	}

	timerDuration := timeToDeletion
	deleleTimerMs := os.Getenv("OW_DELETE_DURATION")
	if deleleTimerMs != "" {
		dur, err := time.ParseDuration(deleleTimerMs)
		if err != nil {
			Debug("Error parsing OW_DELETE_DURATION: %v", err)
		} else {
			timerDuration = dur
		}
	}

	Debug("Starting wait cycle for stopping hash '%s'", actionCodeHash)
	<-time.After(timerDuration)
	Debug("Ended wait cycle for stopping hash '%s'", actionCodeHash)

	if len(innerAPValue.connectedActionIDs) == 0 {
		stopAndDelete(serverAp, innerAPValue, actionCodeHash)
		return
	}

	Debug("Stopping request for hash '%s' skipped, as new actions joined.", actionCodeHash)
}
