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
	"time"

	"github.com/stretchr/testify/require"
)

func TestStopHandler(t *testing.T) {
	t.Run("clean up action", func(t *testing.T) {
		oldCurrentDir, _ := os.Getwd()

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
			actions: make(map[RemoteAPKey]*RemoteAPValue),
		}

		ts := httptest.NewServer(rootAP)

		dat, _ := os.ReadFile("_test/hello_message")
		enc := base64.StdEncoding.EncodeToString(dat)
		body := initBodyRequest{Binary: true, Code: enc}

		actionCodeHash := calculateCodeHash(enc)
		body.Env = map[string]interface{}{OW_CODE_HASH: actionCodeHash}

		initBody, _ := json.Marshal(initRequest{Value: body, ProxiedActionID: actionID})
		doInit(ts, string(initBody))
		require.Contains(t, rootAP.serverProxyData.actions, actionCodeHash)
		lastAction := highestDir(dir)
		require.Greater(t, lastAction, 0)

		doStop(ts, actionCodeHash, actionID)

		require.NotContains(t, rootAP.serverProxyData.actions, actionCodeHash)
		require.NoDirExistsf(t, filepath.Join(dir, strconv.Itoa(lastAction)), "lastAction dir should be removed")
		require.DirExists(t, dir)

		os.RemoveAll(dir)

		stopTestServer(ts, oldCurrentDir, buf)
	})

	t.Run("don't clean up action when other clients still present", func(t *testing.T) {
		oldCurrentDir, _ := os.Getwd()

		actionID := "test-action-id"
		otherActionID := "other-action-id"
		// temporary workdir
		dir, _ := os.MkdirTemp("", "action")
		file, _ := filepath.Abs("_test")
		os.Symlink(file, dir+"/_test")
		os.Chdir(dir)

		// setup the server
		buf, _ := os.CreateTemp("", "log")
		rootAP := NewActionProxy(dir, "", buf, buf, ProxyModeServer)
		rootAP.serverProxyData = &ServerProxyData{
			actions: make(map[RemoteAPKey]*RemoteAPValue),
		}

		ts := httptest.NewServer(rootAP)

		dat, _ := os.ReadFile("_test/hello_message")
		enc := base64.StdEncoding.EncodeToString(dat)
		body := initBodyRequest{Binary: true, Code: enc}

		actionCodeHash := calculateCodeHash(enc)
		body.Env = map[string]interface{}{OW_CODE_HASH: actionCodeHash}

		// first client
		initBody, _ := json.Marshal(initRequest{Value: body, ProxiedActionID: actionID})
		doInit(ts, string(initBody))

		// second client
		initBody, _ = json.Marshal(initRequest{Value: body, ProxiedActionID: otherActionID})
		doInit(ts, string(initBody))

		require.Contains(t, rootAP.serverProxyData.actions, actionCodeHash)

		lastAction := highestDir(dir)
		require.Greater(t, lastAction, 0)

		doStop(ts, actionCodeHash, actionID)

		require.Contains(t, rootAP.serverProxyData.actions, actionCodeHash)
		require.DirExistsf(t, filepath.Join(dir, strconv.Itoa(lastAction)), "lastAction dir should still exist")
		require.DirExists(t, dir)

		os.RemoveAll(dir)

		stopTestServer(ts, oldCurrentDir, buf)
	})

	t.Run("don't clean up action when special action setup is still running", func(t *testing.T) {
		oldCurrentDir, _ := os.Getwd()

		oldSetupActionPath := setupActionPath
		tmpDir := t.TempDir()
		setupActionPath = tmpDir

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
			actions: make(map[RemoteAPKey]*RemoteAPValue),
		}

		ts := httptest.NewServer(rootAP)

		dat, _ := os.ReadFile("_test/hello_message")
		enc := base64.StdEncoding.EncodeToString(dat)
		body := initBodyRequest{Binary: true, Code: enc}

		actionCodeHash := calculateCodeHash(enc)
		body.Env = map[string]interface{}{OW_CODE_HASH: actionCodeHash}

		initBody, _ := json.Marshal(initRequest{Value: body, ProxiedActionID: actionID})
		doInit(ts, string(initBody))
		require.Contains(t, rootAP.serverProxyData.actions, actionCodeHash)
		lastAction := highestDir(dir)
		require.Greater(t, lastAction, 0)

		setupCheckFile := filepath.Join(setupActionPath, actionCodeHash)
		err := os.WriteFile(setupCheckFile, []byte("setup"), 0644)
		require.NoError(t, err)

		doStop(ts, actionCodeHash, actionID)

		require.Contains(t, rootAP.serverProxyData.actions, actionCodeHash)
		require.Empty(t, rootAP.serverProxyData.actions[actionCodeHash].connectedActionIDs)
		require.DirExistsf(t, filepath.Join(dir, strconv.Itoa(lastAction)), "lastAction dir should not be removed during setup")
		require.DirExists(t, dir)

		os.RemoveAll(dir)
		os.Remove(setupCheckFile)

		stopTestServer(ts, oldCurrentDir, buf)
		setupActionPath = oldSetupActionPath
	})

	t.Run("clean up action when special action setup is completed", func(t *testing.T) {
		oldCurrentDir, _ := os.Getwd()

		oldSetupActionPath := setupActionPath
		tmpDir := t.TempDir()
		setupActionPath = tmpDir

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
			actions: make(map[RemoteAPKey]*RemoteAPValue),
		}

		ts := httptest.NewServer(rootAP)

		dat, _ := os.ReadFile("_test/hello_message")
		enc := base64.StdEncoding.EncodeToString(dat)
		body := initBodyRequest{Binary: true, Code: enc}

		actionCodeHash := calculateCodeHash(enc)
		body.Env = map[string]interface{}{OW_CODE_HASH: actionCodeHash}

		initBody, _ := json.Marshal(initRequest{Value: body, ProxiedActionID: actionID})
		doInit(ts, string(initBody))
		require.Contains(t, rootAP.serverProxyData.actions, actionCodeHash)
		lastAction := highestDir(dir)
		require.Greater(t, lastAction, 0)

		setupCheckFile := filepath.Join(setupActionPath, actionCodeHash)
		setupDoneFile := setupCheckFile + "_done"
		err := os.WriteFile(setupCheckFile, []byte("setup"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(setupDoneFile, []byte("done"), 0644)
		require.NoError(t, err)

		doStop(ts, actionCodeHash, actionID)

		require.NotContains(t, rootAP.serverProxyData.actions, actionCodeHash)
		require.NoDirExistsf(t, filepath.Join(dir, strconv.Itoa(lastAction)), "lastAction dir should be removed")
		require.DirExists(t, dir)

		os.RemoveAll(dir)

		stopTestServer(ts, oldCurrentDir, buf)
		setupActionPath = oldSetupActionPath
	})

	t.Run("clean up action after timer is done", func(t *testing.T) {
		oldCurrentDir, _ := os.Getwd()

		oldSetupActionPath := setupActionPath
		tmpDir := t.TempDir()
		setupActionPath = tmpDir

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
			actions: make(map[RemoteAPKey]*RemoteAPValue),
		}

		ts := httptest.NewServer(rootAP)

		dat, _ := os.ReadFile("_test/hello_message")
		enc := base64.StdEncoding.EncodeToString(dat)
		body := initBodyRequest{Binary: true, Code: enc}

		actionCodeHash := calculateCodeHash(enc)
		body.Env = map[string]interface{}{OW_CODE_HASH: actionCodeHash}

		initBody, _ := json.Marshal(initRequest{Value: body, ProxiedActionID: actionID})
		doInit(ts, string(initBody))
		require.Contains(t, rootAP.serverProxyData.actions, actionCodeHash)
		lastAction := highestDir(dir)
		require.Greater(t, lastAction, 0)

		setupCheckFile := filepath.Join(setupActionPath, actionCodeHash)
		err := os.WriteFile(setupCheckFile, []byte("setup"), 0644)
		require.NoError(t, err)

		oldtimerToDeletion := timeToDeletion
		timeToDeletion = 100 * time.Millisecond
		doStop(ts, actionCodeHash, actionID)

		require.Contains(t, rootAP.serverProxyData.actions, actionCodeHash)

		time.Sleep(110 * time.Millisecond)

		require.NotContains(t, rootAP.serverProxyData.actions, actionCodeHash)
		require.NoDirExistsf(t, filepath.Join(dir, strconv.Itoa(lastAction)), "lastAction dir should be removed")
		require.DirExists(t, dir)

		os.RemoveAll(dir)

		stopTestServer(ts, oldCurrentDir, buf)

		timeToDeletion = oldtimerToDeletion
		setupActionPath = oldSetupActionPath
	})
}
func TestIsSetupActionRunning(t *testing.T) {
	oldSetupActionPath := setupActionPath

	t.Run("setup action is running", func(t *testing.T) {
		tmpDir := t.TempDir()
		setupActionPath = tmpDir
		actionCodeHash := "test-action-hash"
		setupCheckFile := filepath.Join(setupActionPath, actionCodeHash)

		// Create the setup check file
		err := os.WriteFile(setupCheckFile, []byte("setup"), 0644)
		require.NoError(t, err)

		// Check if setup action is running
		running := isSetupActionRunning(actionCodeHash)
		require.True(t, running)
	})

	t.Run("setup action is not running", func(t *testing.T) {
		tmpDir := t.TempDir()
		setupActionPath = tmpDir
		actionCodeHash := "test-action-hash"

		// Check if setup action is running
		running := isSetupActionRunning(actionCodeHash)
		require.False(t, running)
	})

	t.Run("setup action is completed", func(t *testing.T) {
		tmpDir := t.TempDir()
		setupActionPath = tmpDir
		actionCodeHash := "test-action-hash"
		setupCheckFile := filepath.Join(setupActionPath, actionCodeHash)
		setupDoneFile := setupCheckFile + "_done"

		// Create the setup check file
		err := os.WriteFile(setupCheckFile, []byte("setup"), 0644)
		require.NoError(t, err)

		// Create the setup done file
		err = os.WriteFile(setupDoneFile, []byte("done"), 0644)
		require.NoError(t, err)

		// Check if setup action is running
		running := isSetupActionRunning(actionCodeHash)
		require.False(t, running)
	})

	setupActionPath = oldSetupActionPath
}
