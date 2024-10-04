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
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStopHandler(t *testing.T) {

	actionID := "test-action-id"
	// temporary workdir
	dir, _ := os.MkdirTemp("", "action")
	file, _ := filepath.Abs("_test")
	os.Symlink(file, dir+"/_test")
	os.Chdir(dir)
	// setup the server
	buf, _ := os.CreateTemp("", "log")
	rootAP := NewActionProxy(dir, "", buf, buf, ProxyModeServer)
	rootAP.serverProxyData = &ServerProxyData{
		actions: make(map[string]*ActionProxy),
	}

	ts := httptest.NewServer(rootAP)

	dat, _ := os.ReadFile("_test/hello_message")
	enc := base64.StdEncoding.EncodeToString(dat)
	body := initBodyRequest{Binary: true, Code: enc}
	initBody, _ := json.Marshal(initRequest{Value: body, ProxiedActionID: actionID})
	doInit(ts, string(initBody))
	require.Contains(t, rootAP.serverProxyData.actions, actionID)
	lastAction := highestDir(dir)

	doRun(ts, "")
	doStop(ts, actionID)

	require.NotContains(t, rootAP.serverProxyData.actions, actionID)
	require.NoDirExistsf(t, filepath.Join(dir, strconv.Itoa(lastAction)), "lastAction dir should be removed")

}
