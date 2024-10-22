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
	"net/http"
)

func (ap *ActionProxy) resetHandler(w http.ResponseWriter, _ *http.Request) {
	if ap.proxyMode != ProxyModeServer {
		sendError(w, http.StatusMethodNotAllowed, "Reset allowed only in server mode")
		return
	}

	for _, ap := range ap.serverProxyData.actions {
		cleanUpAP(ap)
	}

	ap.serverProxyData.actions = make(map[string]*ActionProxy)

	sendOK(w)
}