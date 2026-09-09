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

#ARG COMMON=missing:missing
ARG COMMON=registry.hub.docker.com/apache/openserverless-runtime-common:common1.18.4
FROM ${COMMON} AS builder

FROM python:3.12.2-bookworm
ENV OW_EXECUTION_ENV=apacheopenserverless/runtime-python-v3.12.2
COPY --from=builder /go/bin/proxy /bin/proxy 

# install zip
RUN apt-get update && apt-get install -y python3-lxml python3-psycopg2 zip xpdf \
    && rm -rf /var/lib/apt/lists/*

# Install common modules for python
COPY requirements.txt requirements.txt

RUN pip3 install --upgrade pip six wheel virtualenv
RUN pip3 install --no-cache-dir -r requirements.txt

RUN mkdir -p /action
WORKDIR /

COPY bin/compile /bin/compile
COPY lib/launcher.py /lib/launcher.py

# log initialization errors
ENV OW_LOG_INIT_ERROR=1
# the launcher must wait for an ack
ENV OW_WAIT_FOR_ACK=1
# using the runtime name to identify the execution environment
# compiler script
ENV OW_COMPILER=/bin/compile
# home directory
ENV HOME=/tmp

WORKDIR /action
RUN chown nobody:root /action ; chmod 0775 /action
USER nobody
ENTRYPOINT ["/bin/proxy"]
