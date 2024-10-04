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
	"os/signal"
	"strconv"
	"syscall"
)

func (ap *ActionProxy) HookExitSignals() {
	signalChan := make(chan os.Signal, 1)
	listenOnExitSignals(ap, signalChan)
	os.Exit(0)
}

func listenOnExitSignals(ap *ActionProxy, captureSignalChan chan os.Signal) {
	// Relay stop signals to captureSignalChan
	signal.Notify(captureSignalChan,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGABRT,
		syscall.SIGQUIT,
		syscall.SIGHUP)
	defer signal.Stop(captureSignalChan)

	Debug("Listening on exit signals for remote action cleanup...")
	signalHandler(<-captureSignalChan, ap)
}

func signalHandler(signal os.Signal, ap *ActionProxy) {
	Debug("Caught signal: %v", signal)

	_ = ap.SendStopRequest()

	Debug("Finished remote action cleanup. Exiting.")
}

func (ap *ActionProxy) SendStopRequest() error {
	if ap.clientProxyData == nil {
		Debug("Nothing to stop")
		return fmt.Errorf("runtime not set as client")
	}

	stopRequest := stopRequest{
		ProxiedActionID: ap.clientProxyData.ProxyActionID,
	}

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(stopRequest)
	if err != nil {
		Debug("Failed to send stop request: error encoding stop request body: %v", err)
		return err
	}

	bodyLen := buf.Len()

	body := io.NopCloser(bytes.NewBuffer(buf.Bytes()))
	url := ap.clientProxyData.ProxyURL.String() + "/stop"
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		Debug("Failed to send stop request: error creating stop request: %v", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", strconv.Itoa(bodyLen))
	req.ContentLength = int64(bodyLen)

	client := &http.Client{}

	Debug("Sending stop request to %s", url)
	resp, err := client.Do(req)
	if err != nil {
		Debug("Failed to send stop request: %v", err)
		return err
	}

	respBody, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	Debug("Stop request response: %v", string(respBody))
	return nil
}
