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
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCodeHashForwarded(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var initRequest initRequest
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		err = json.Unmarshal(body, &initRequest)
		require.NoError(t, err)

		require.Contains(t, initRequest.Value.Env, OW_CODE_HASH)

	}))

	// create a client ActionProxy
	clientAP := NewActionProxy("", "", nil, nil, ProxyModeClient)

	// create a request body
	body := initBinary("_test/hello.zip", "@"+ts.URL)
	initBody := bytes.NewBufferString(body)
	w := httptest.NewRecorder()

	// run the client init request
	clientAP.initHandler(w, httptest.NewRequest(http.MethodPost, "/", initBody))

	// stop the server
	runtime.Gosched()
	// wait 2 seconds before declaring a test done
	time.Sleep(2 * time.Second)
	ts.Close()
}

func Example_forwardInitRequest() {
	// create a client ActionProxy
	clientAP := NewActionProxy("", "", nil, nil, ProxyModeClient)

	// create a server ActionProxy
	compiler, _ := filepath.Abs("common/gobuild.py")
	log, _ := os.CreateTemp("", "log")
	serverAP := NewActionProxy("./action", compiler, log, log, ProxyModeServer)

	// start the server
	ts := httptest.NewServer(serverAP)

	// create a request body
	body := initBinary("_test/hello.zip", "@"+ts.URL)
	initBody := bytes.NewBufferString(body)
	w := httptest.NewRecorder()

	// run the client init request
	clientAP.initHandler(w, httptest.NewRequest(http.MethodPost, "/init", initBody))

	// stop the server
	runtime.Gosched()
	// wait 2 seconds before declaring a test done
	time.Sleep(2 * time.Second)
	ts.Close()
	dump(log)

	fmt.Print(w.Body.String())
	fmt.Println(clientAP.clientProxyData.ActionCodeHash != "")

	// Output:
	// {"ok":true}
	// true
}

func Example_forwardRunRequest() {
	clientLog, _ := os.CreateTemp("", "log")
	// create a client ActionProxy
	clientAP := NewActionProxy("", "", clientLog, clientLog, ProxyModeClient)

	// create a server ActionProxy
	compiler, _ := filepath.Abs("common/gobuild.py")
	serverAP := NewActionProxy("./action", compiler, nil, nil, ProxyModeServer)

	// start the server
	ts := httptest.NewServer(serverAP)

	// create a request body
	body := initBinary("_test/hello.zip", "@"+ts.URL)
	initBody := bytes.NewBufferString(body)
	w := httptest.NewRecorder()

	// run the client init request
	clientAP.initHandler(w, httptest.NewRequest(http.MethodPost, "/init", initBody))
	fmt.Print(w.Body.String())

	// create a request body
	runW := httptest.NewRecorder()
	runBody := bytes.NewBufferString(`{"value": {"name": "Mike"}}`)
	clientAP.runHandler(runW, httptest.NewRequest(http.MethodPost, "/run", runBody))

	// stop the server
	runtime.Gosched()
	// wait 2 seconds before declaring a test done
	time.Sleep(2 * time.Second)
	ts.Close()
	dump(clientLog)
	fmt.Println(runW.Body.String())
	os.Remove(clientLog.Name())

	// Output:
	// {"ok":true}
	// Main
	// Hello, Mike
	// XXX_THE_END_OF_A_WHISK_ACTIVATION_XXX
	// XXX_THE_END_OF_A_WHISK_ACTIVATION_XXX
	// {"greetings":"Hello, Mike"}
}

func TestParseMainFlag(t *testing.T) {
	tests := []struct {
		name          string
		mainAtProxy   string
		expectedMain  string
		expectedProxy string
		expectError   bool
	}{
		{
			name:          "Valid main and proxy",
			mainAtProxy:   "mainFunc@https://example.com",
			expectedMain:  "mainFunc",
			expectedProxy: "https://example.com",
			expectError:   false,
		},
		{
			name:          "Valid proxy only",
			mainAtProxy:   "https://example.com",
			expectedMain:  "",
			expectedProxy: "https://example.com",
			expectError:   false,
		},
		{
			name:        "Invalid main flag format",
			mainAtProxy: "mainFunc@https://example.com@extra",
			expectError: true,
		},
		{
			name:        "Invalid URL",
			mainAtProxy: "mainFunc@://invalid-url",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxyData, err := parseMainFlag(tt.mainAtProxy)

			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expectedMain, proxyData.MainFunc)
			require.Equal(t, tt.expectedProxy, proxyData.ProxyURL.String())
		})
	}
}
func TestParseMainURL(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedURL string
		expectError bool
	}{
		{
			name:        "Valid URL with scheme",
			input:       "https://example.com",
			expectedURL: "https://example.com",
			expectError: false,
		},
		{
			name:        "Valid URL without scheme",
			input:       "example.com",
			expectedURL: "https://example.com",
			expectError: false,
		},
		{
			name:        "Invalid URL",
			input:       "://invalid-url",
			expectError: true,
		},
		{
			name:        "Empty URL",
			input:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedURL, err := parseMainURL(tt.input)
			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expectedURL, parsedURL.String())
		})
	}
}
