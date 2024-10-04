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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSendStopRequest(t *testing.T) {

	t.Run("clientProxyData is nil", func(t *testing.T) {
		ap := &ActionProxy{}
		err := ap.SendStopRequest()
		require.Error(t, err)
		require.Contains(t, err.Error(), "runtime not set as client")
	})

	t.Run("success", func(t *testing.T) {
		mockedServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var stopReq stopRequest
			err := json.NewDecoder(r.Body).Decode(&stopReq)
			require.NoError(t, err)
			require.Equal(t, "test-action-id", stopReq.ProxiedActionID)
		}))

		url, _ := url.Parse(mockedServer.URL)
		ap := &ActionProxy{
			clientProxyData: &ClientProxyData{
				ProxyActionID: "test-action-id",
				ProxyURL:      *url,
			},
		}

		err := ap.SendStopRequest()
		require.NoError(t, err)
	})
}
func TestHookExitSignals(t *testing.T) {
	mockedServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var stopReq stopRequest
		err := json.NewDecoder(r.Body).Decode(&stopReq)
		require.NoError(t, err)
		require.Equal(t, "test-action-id", stopReq.ProxiedActionID)
	}))

	url, _ := url.Parse(mockedServer.URL)

	ap := &ActionProxy{
		clientProxyData: &ClientProxyData{
			ProxyActionID: "test-action-id",
			ProxyURL:      *url,
		},
	}

	signalChan := make(chan os.Signal, 1)
	signalChan <- os.Interrupt
	ap.HookExitSignals(signalChan)
}
