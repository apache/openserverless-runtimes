<!--
#
# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
-->

# Apache OpenServerless Runtimes

All the Apache Openserverless OpenWhisk runtimes in a single place using the Go proxy and ActionLoop.

# Source Code

runtimes are docker images, and they all use a proxy in go and some scripts for execution.

Go Proxy code is in folder `openwhisk` and the main is `proxy.go` in top level.

You can compile it with `go build -o proxy`.

Tests are in openwhisk folder, test it with `cd opewhisk ; go test `

Runtime sources are under `runtimes/<plang>/<version>` (`<plang>` is programming languate)

Special case is `runtime/common/<version>` that contains the proxy itseself, it is used as base image for the others and must be build first.

# How to build images

Build and push the common runtime with `task build-common`. Also ensure the image is public.

Then you can build a single runtime specifingh the dir:

Build a single runtime: `task build-runtime RT=nodejs VER=v18`

# How to generate a new runtimes.json

The project contains a `runtimes.json.tpl` with specific placeholder for managed Apache OpenServerless runtimes. To regenerate a newer `runtimes.json` from the current TAG
and assuming that the images have been effectively pushed to the Apache Official DockerHub repositories execute the command:

`task render-runtimes`

This will create a new `runtimes.json` that can be pushed to the official Apache OpenServerless [task](https://github.com/apache/openserverless-task) repo, replacing the existing file.

# How to use the client/server mode

The proxy can be used in client or server mode, where the client acts as a forward
proxy and the server will be the actual executor.

In client mode the runtime does not execute the action, but instead forwards the
/init and /run requests to a server runtime. To activate this mode, set the environment
variable `OW_ACTIVATE_PROXY_CLIENT` to 1.
When creating actions, use the --main flag with this syntax:
`--main "<main>@<remote runtime address>"`. `<main>` can be empty.

The remote runtime is enabled by setting the environment variable `OW_ACTIVATE_PROXY_SERVER` to 1.
In this mode the runtime is multi-action enabled, meaning that it can initialize and run more than one action.
Many client runtimes can forward requests to the same server runtime.

Currently the proxy client/server extension has been compiled and releaed inside the common runtime `common1.18.0`. 

The go runtimes have been extended with the runtime `v1.22proxy` which has been setup by default with the `OW_ACTIVATE_PROXY_CLIENT` set to 1. To deploy an action to be proxid remotely
use a command similar to `ops action create <action> tests/pytorch.py --main main@http://ops-cuda-service:8080 --kind go:1.22proxy`

The experimental runtime to be used as remote proxy server are currently within the `runtime\experimental` folder and can be built using `task build-experimental-runtimes`. These runtime have to be deployed as regular pod/container on a remote machine. To activate the proxy server mode endure that the image is launched setting the environment variable `OW_ACTIVATE_PROXY_SERVER=1`, otherwise the runtime behaves as a regular OpenWhisk one.
