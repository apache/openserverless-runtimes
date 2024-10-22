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
	"path/filepath"
	"strconv"
)

type stopRequest struct {
	ProxiedActionID string `json:"proxiedActionID,omitempty"`
}

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

	actionID := stopRequest.ProxiedActionID

	innerAP, ok := ap.serverProxyData.actions[actionID]
	if !ok {
		Debug("Action '%s' not found in server proxy data", actionID)
		sendError(w, http.StatusNotFound, "Action to be removed in remote runtime not found. Check logs for details.")
	}

	Debug("Action '%s' executor stopped", actionID)
	cleanUpAP(innerAP)
	delete(ap.serverProxyData.actions, actionID)

	sendOK(w)
}

func cleanUpAP(ap *ActionProxy) {
	ap.theExecutor.Stop()
	if err := os.RemoveAll(filepath.Join(ap.baseDir, strconv.Itoa(ap.currentDir))); err != nil {
		Debug("Error removing action directory: %v", err)
	}
}
