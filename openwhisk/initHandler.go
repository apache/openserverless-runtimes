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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type initBodyRequest struct {
	Code   string                 `json:"code,omitempty"`
	Binary bool                   `json:"binary,omitempty"`
	Main   string                 `json:"main,omitempty"`
	Env    map[string]interface{} `json:"env,omitempty"`
}

type initRequest struct {
	ProxiedActionID string          `json:"proxiedActionID,omitempty"`
	Value           initBodyRequest `json:"value,omitempty"`
}

func sendOK(w http.ResponseWriter) {
	// answer OK
	w.Header().Set("Content-Type", "application/json")
	buf := []byte("{\"ok\":true}\n")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(buf)))
	w.Write(buf)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

func (ap *ActionProxy) initHandler(w http.ResponseWriter, r *http.Request) {
	// you can do multiple initializations when debugging
	if ap.initialized && !Debugging {
		msg := "Cannot initialize the action more than once."
		sendError(w, http.StatusForbidden, msg)
		log.Println(msg)
		return
	}

	if ap.proxyMode == ProxyModeClient {
		ap.ForwardInitRequest(w, r)
		return
	}

	if ap.compiler != "" {
		Debug("compiler: " + ap.compiler)
	}

	// read body of the request
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
		return
	}

	// decode request parameters
	if len(body) < 1000 {
		Debug("init: decoding %s\n", string(body))
	}

	var request initRequest
	err = json.Unmarshal(body, &request)
	if err != nil {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("Error unmarshaling request: %v", err))
		return
	}

	if ap.proxyMode == ProxyModeServer {
		if ap.serverProxyData == nil {
			ap.serverProxyData = &ServerProxyData{actions: make(map[RemoteAPKey]*RemoteAPValue)}
		}
		if ok := doRemoteInit(ap, request, w); !ok {
			return
		}

		sendOK(w)
		return
	}

	if err := ap.doInit(request, w); err != nil {
		Debug("Error initializing action: %v", err)
		return
	}

	sendOK(w)
}

// doRemoteInit initializes a remote action.
// Returns true if the initialization was successful, false otherwise.
func doRemoteInit(ap *ActionProxy, request initRequest, w http.ResponseWriter) bool {
	Debug("Remote initialization started.")

	// Get the action code hash from the client request
	actionCodeHash, ok := request.Value.Env[OW_CODE_HASH].(string)
	if !ok {
		sendError(w, http.StatusBadGateway, "Cannot identify the action in remote runtime (missing hash).")
		return false
	}

	if request.ProxiedActionID == "" {
		sendError(w, http.StatusBadGateway, "Missing action id from client.")
		return false
	}

	Debug("Action code hash extracted: %s", actionCodeHash)

	// check if the action is already initialized
	if nestedAP, ok := ap.serverProxyData.actions[actionCodeHash]; ok {
		Debug("Action already initialized. Added action ID %s to action hash %s", request.ProxiedActionID, actionCodeHash)
		nestedAP.connectedActionIDs = append(nestedAP.connectedActionIDs, request.ProxiedActionID)
		sendOK(w)
		return true
	}

	outLog, err := os.CreateTemp("", "out-log")
	if err != nil {
		outLog = ap.outFile
	}
	errLog, err := os.CreateTemp("", "err-log")
	if err != nil {
		errLog = ap.errFile
	}

	Debug("Creating nested action proxy...")
	innerActionProxy := NewActionProxy(ap.baseDir, ap.compiler, outLog, errLog, ProxyModeNone)
	if err := innerActionProxy.doInit(request, w); err != nil {
		return false
	}

	ap.serverProxyData.actions[actionCodeHash] = &RemoteAPValue{
		remoteProxy:        innerActionProxy,
		connectedActionIDs: []string{request.ProxiedActionID},
		runRequestQueue:    make(chan *remoteRunChanPayload, 50), // size could be determined empirically
	}

	Debug("Started listening to run requests for AP with code hash %s...", actionCodeHash)
	go startListenToRunRequests(innerActionProxy, ap.serverProxyData.actions[actionCodeHash].runRequestQueue)

	Debug("Added action id %s to action hash %s", request.ProxiedActionID, actionCodeHash)
	return true
}

func (ap *ActionProxy) doInit(request initRequest, w http.ResponseWriter) error {
	// request with empty code - stop any executor but return ok
	if request.Value.Code == "" {
		sendError(w, http.StatusForbidden, "Missing main/no code to execute.")
		return fmt.Errorf("code in body is empty")
	}

	// passing the env to the action proxy
	ap.SetEnv(request.Value.Env)

	// setting main
	main := request.Value.Main
	if main == "" {
		main = "main"
	}

	// extract code eventually decoding it
	var buf []byte
	if request.Value.Binary {
		Debug("it is binary code")
		b, err := base64.StdEncoding.DecodeString(request.Value.Code)
		if err != nil {
			sendError(w, http.StatusBadRequest, "cannot decode the request: "+err.Error())
			return err
		}
		buf = b
	} else {
		Debug("it is source code")
		buf = []byte(request.Value.Code)
	}

	// if a compiler is defined try to compile
	_, err := ap.ExtractAndCompile(&buf, main)
	if err != nil {
		if os.Getenv("OW_LOG_INIT_ERROR") == "" {
			sendError(w, http.StatusBadGateway, err.Error())
		} else {
			ap.errFile.Write([]byte(err.Error() + "\n"))
			ap.outFile.Write([]byte(OutputGuard))
			ap.errFile.Write([]byte(OutputGuard))
			sendError(w, http.StatusBadGateway, "The action failed to generate or locate a binary. See logs for details.")
		}
		return err
	}

	// start an action
	err = ap.StartLatestAction()
	if err != nil {
		if os.Getenv("OW_LOG_INIT_ERROR") == "" {
			sendError(w, http.StatusBadGateway, "cannot start action: "+err.Error())
		} else {
			ap.errFile.Write([]byte(err.Error() + "\n"))
			ap.outFile.Write([]byte(OutputGuard))
			ap.errFile.Write([]byte(OutputGuard))
			sendError(w, http.StatusBadGateway, "Cannot start action. Check logs for details.")
		}
		return err
	}
	ap.initialized = true

	return nil
}

// ExtractAndCompile decode the buffer and if a compiler is defined, compile it also
func (ap *ActionProxy) ExtractAndCompile(buf *[]byte, main string) (string, error) {

	// extract action in src folder
	file, err := ap.ExtractAction(buf, "src")
	if err != nil {
		return "", err
	}
	if file == "" {
		return "", fmt.Errorf("empty filename")
	}

	// some path surgery
	dir := filepath.Dir(file)
	parent := filepath.Dir(dir)
	srcDir := filepath.Join(parent, "src")
	binDir := filepath.Join(parent, "bin")
	binFile := filepath.Join(binDir, "exec")

	// if the file is already compiled or there is no compiler just move it from src to bin
	if ap.compiler == "" || isCompiled(file) {
		os.Rename(srcDir, binDir)
		return binFile, nil
	}

	// ok let's try to compile
	Debug("compiling: %s main: %s", file, main)
	os.Mkdir(binDir, 0755)
	err = ap.CompileAction(main, srcDir, binDir)
	if err != nil {
		return "", err
	}

	// check only if the file exist
	if _, err := os.Stat(binFile); os.IsNotExist(err) {
		return "", fmt.Errorf("cannot compile")
	}
	return binFile, nil
}
