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

# Do not fix the patch level for golang:1.20 to automatically get security fixes.
ARG COMMON=missing:missing
FROM ${COMMON} AS builder
FROM golang:1.26.5-bookworm AS build
COPY --from=builder /go/bin/proxy /bin/proxy

# Install delve in build stage
RUN go install github.com/go-delve/delve/cmd/dlv@latest

# Final runtime stage
FROM golang:1.26.5-bookworm
ENV OW_EXECUTION_ENV=apacheopenserverless/runtime-golang-v1.26.5

# Copy proxy and delve from build stage
COPY --from=build /bin/proxy /bin/proxy
COPY --from=build /go/bin/dlv /go/bin/dlv

# Install only runtime dependencies
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    curl \
    jq \
    git \
    zip \
    librdkafka1 \
    librdkafka++1 && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* && \
    mkdir /action /home/go

COPY bin/compile /bin/compile
COPY lib/launcher.go /lib/launcher.go
COPY bin/extend  /bin/extend
ENV OW_COMPILER=/bin/compile \
    OW_LOG_INIT_ERROR=1 \
    OW_WAIT_FOR_ACK=1

WORKDIR /action
RUN chown nobody:root /action /home/go && chmod 0775 /action /home/go
USER nobody

ENTRYPOINT [ "/bin/proxy" ]
