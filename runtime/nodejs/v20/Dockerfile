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

FROM node:20.11.1-bookworm
ENV OW_EXECUTION_ENV=apacheopenserverless/nodejs-20.11.1
COPY --from=builder /go/bin/proxy /bin/proxy 

ENV OW_COMPILER=/bin/compile
ENV OW_LOG_INIT_ERROR=1
ENV OW_WAIT_FOR_ACK=1

RUN apt-get update && apt-get install -y \
    imagemagick \
    graphicsmagick \
    unzip \
    && rm -rf /var/lib/apt/lists/*

COPY package.json /
RUN npm config set fetch-retry-maxtimeout 60000 -g
RUN cd / && npm install && npm cache clean --force

RUN mkdir -p /app/action ;\
  chmod -R 0775 /app ;\
  chown -R nobody:root /app

ENV HOME=/app
WORKDIR /app
COPY bin/compile /bin/compile
COPY lib/launcher.js /lib/launcher.js

EXPOSE 8080
USER nobody

ENTRYPOINT ["/bin/proxy"]
