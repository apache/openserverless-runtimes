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
# Nuvolaris Runtines Next Generation

All the runtimes in a single place using the Go proxy and ActionLoop.

# Source Code

runtimes are docker images, and they all use a proxy in go and some scripts for execution.

Go Proxy code is in folder `openwhisk` and the main is `proxy.go` in top level.

You can  compile it with `go build -o proxy`. 

Tests are in openwhisk folder, test it with `cd opewhisk ; go test `

Runtime  sources are under `runtimes/<plang>/<version>` (`<plang>` is programming languate)

Special case is `runtime/common/<version>` that contains the proxy itseself, it is used as base image for the others and must be build first.

# How to build images

Build and push the common runtime  with `task build-common`. Also ensure the image is public.

Then you can build a single runtime specifingh the dir:

Build a single runtime: `task build-runtime DIR=nodejs/v18`