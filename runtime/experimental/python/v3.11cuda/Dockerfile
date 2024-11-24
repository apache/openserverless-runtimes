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

# This runtime is using the experimental proxy common1.18.0
ARG COMMON=missing:missing
FROM ${COMMON} AS builder

FROM pytorch/pytorch:2.5.1-cuda12.4-cudnn9-runtime
ENV OW_EXECUTION_ENV=apacheopenserverless/runtime-python-v3.11
COPY --from=builder /go/bin/proxy /bin/proxy

# install zip
RUN apt-get update && apt-get install -y zip xpdf\
    build-essential g++ ninja-build \
    && rm -rf /var/lib/apt/lists/*  

COPY requirements.txt requirements.txt

RUN \
    /opt/conda/bin/conda init ;\
    source /opt/conda/bin/activate ;\
    conda install lxml psycopg2 -y ;\
    pip3 install --upgrade pip six wheel virtualenv ;\
    pip3 install --no-cache-dir -r requirements.txt

RUN mkdir -p /action
WORKDIR /

ADD bin/compile /bin/compile
ADD lib/launcher.py /lib/launcher.py
ADD debug.sh /bin/debug.sh

# log initialization errors
ENV OW_LOG_INIT_ERROR=1
# the launcher must wait for an ack
ENV OW_WAIT_FOR_ACK=1
# using the runtime name to identify the execution environment
# compiler script
ENV OW_COMPILER=/bin/compile

# set this runtime as a remote server (this could be passed as ENV variable at deployment time)
# ENV OW_ACTIVATE_PROXY_SERVER=1

# home directory

ENV HOME=/workspace
WORKDIR /action
RUN mkdir -p /action ;\
    chmod 0775 /action /workspace ;\
    useradd -m -d /workspace -s /bin/bash -g root workspace ;\
    chown workspace:root /action /workspace 

ENTRYPOINT ["/bin/proxy"]
