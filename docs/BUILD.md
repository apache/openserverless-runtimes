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
#  Developers Guide for Runtime itself

<a name="building"/>

## Prerequisites

You need a Linux or an OSX environment with:

- [Task](https://taskfile.dev) (the build driver, all commands below are `task` targets)
- Docker with `buildx` enabled (images are built for `linux/amd64` and `linux/arm64`)
- Go (to compile the proxy)
- `git` (the build reads the current git tag)

Run `task --list-all` to see every available target.

## How to build the images

Building is a **two step** process:

1. `task tag` — creates a git tag that *selects what to build* and stamps the image version.
2. `task build` — builds exactly what the tag selected.

You always have to run `task tag` first: `task build` does nothing useful without it,
because it derives both the set of runtimes to build and the image tag from the current
git tag.

### Step 1: select what to build with `task tag`

```
task tag
```

With no arguments this tags `all_<timestamp>`, meaning *build everything*.

To restrict the build to a single set of runtimes, pass `RT=<language>`:

```
task tag RT=python
```

`RT` accepts:

- a **language**: one of the runtimes listed in the `RUNTIMES` variable of the `Taskfile.yml`
  (currently `go`, `nodejs`, `python`, `java`), to build only that language's images;
- `common`: to rebuild only the common base image;
- `experimental`: to rebuild only the experimental runtimes (see `TaskfileExperimental.yml`);
- nothing (or `all`): to build the common image and all the language runtimes.

The tag has the form `<RT>_<timestamp><SUFFIX>` (the default `SUFFIX` is `-SNAPSHOT`),
and the command prints the tag it created. Note that `task tag` **removes the existing
local tags** before creating the new one.

### Step 2: build with `task build`

```
task build
```

This reads the current git tag and:

- `all_*` → builds the common image, then every language in `RUNTIMES`;
- `common_*` → builds only the common image;
- `<language>_*` → builds only the images of that language (all the `v*` versions
  under `runtime/<language>/`).

The resulting images are named
`<REGISTRY>/<REPO_PATH>/openserverless-runtime-<language>:<version>-<tag>` and are
loaded in the local Docker daemon.

### Building and pushing

To build multi-architecture images and push them to the registry:

```
task buildx
```

which is just `task build PUSH=y`. Log in first with `task docker-login`
(it uses `DOCKERHUB_USER` / `DOCKERHUB_TOKEN`, and `DOCKERHUB_REGISTRY` if you are not
using `docker.io`; those can be put in a `.env` file).

Add `DRY=echo` to any build command to just print the commands instead of running them.

### Examples

```
# build everything (common + all languages)
task tag && task build

# build only the python runtimes
task tag RT=python && task build

# rebuild only the common base image
task tag RT=common && task build

# rebuild only the experimental runtimes
task tag RT=experimental && task build

# build all and push multi-arch images
task tag && task buildx
```

## Other useful targets

```
task compile                    # build the go proxy locally
task run RT=python VER=v3.13    # run a built runtime on port 8080
task debug RT=python VER=v3.13  # shell into a runtime with the sources mounted
task invoke J='{"name":"x"}'    # invoke a runtime listening on port 8080
task clean-images               # remove the local openserverless images and prune buildx
task render-runtimes            # generate runtimes.json for the current tag
```

<a name="development"/>

# Local Development

If you want to develop the proxy and run tests natively, you can do it on Linux or OSX.

You need [go](https://golang.org/doc/install) and a set of utilities used in tests:

- bc
- zip

Linux: `apt-get install bc zip`
OSX: `brew install zip`

**NOTE**: Because tests build and cache some binary files, perform a `git clean -fx` and
**do not share folders between linux and osx** because binaries are in different format...

To run the tests:

```
task test
```

or directly:

```
cd openwhisk
go test
```
