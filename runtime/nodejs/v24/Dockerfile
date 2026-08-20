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

ARG COMMON=missing:missing
FROM ${COMMON} AS builder

# Build stage for Node.js dependencies
FROM node:24.11.0-bookworm-slim AS build-env

# Install npm dependencies
WORKDIR /build
COPY package.json .
RUN npm config set fetch-retry-maxtimeout 60000 -g && \
    npm install --omit=dev && \
    npm cache clean --force

# Final minimal runtime
FROM node:24.11.0-bookworm-slim

# Set runtime environment
ENV OW_EXECUTION_ENV=apacheopenserverless/nodejs-24.11.0 \
    HOME=/app \
    OW_LOG_INIT_ERROR=1 \
    OW_WAIT_FOR_ACK=1 \
    OW_COMPILER=/bin/compile

# Install only runtime dependencies with --no-install-recommends
RUN apt-get update && apt-get install -y --no-install-recommends \
    imagemagick graphicsmagick unzip jq && \
    apt-get autoremove -y && \
    rm -rf /var/lib/apt/lists/*

# Copy Node.js modules from builder
COPY --from=build-env /build/node_modules /node_modules
COPY --from=build-env /build/package.json /package.json

# Copy OpenWhisk runtime and proxy binary
COPY --from=builder /go/bin/proxy /bin/proxy
COPY --chmod=755 bin/compile /bin/compile
COPY --chmod=755 bin/extend  /bin/extend
COPY lib/launcher.js /lib/launcher.js

# Prepare /app
WORKDIR /app
RUN mkdir -p /app/action && \
    chown -R nobody:root /app && \
    chmod -R 0775 /app

EXPOSE 8080
USER nobody

ENTRYPOINT ["/bin/proxy"]
