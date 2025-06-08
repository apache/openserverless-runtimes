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

FROM python:3.13.4-slim-bookworm AS build-env

# Set environment for uv installation
ENV UV_CACHE_DIR=/tmp/uv-cache \
    UV_INSTALL_DIR=/usr/local/bin

# Install build tools and install uv
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl ca-certificates build-essential python3-dev && \
    curl -LsSf https://astral.sh/uv/install.sh | sh && \
    apt-get purge -y curl && \
    rm -rf /var/lib/apt/lists/*

# Install Python dependencies
WORKDIR /build
COPY requirements.txt .
RUN uv pip install --python python3 --system six wheel virtualenv
RUN uv pip install --python python3 --system --no-cache-dir -r requirements.txt

# Final minimal runtime
FROM python:3.13.4-slim-bookworm

# Set runtime environment
ENV OW_EXECUTION_ENV=apacheopenserverless/runtime-python-v3.13.4 \
    HOME=/tmp \
    OW_LOG_INIT_ERROR=1 \
    OW_WAIT_FOR_ACK=1 \
    OW_COMPILER=/bin/compile

# Install only runtime deps
RUN apt-get update && apt-get install -y --no-install-recommends \
    python3-psycopg2 zip xpdf ca-certificates && \
    apt-get autoremove -y && \
    rm -rf /var/lib/apt/lists/*

# Copy uv binary and Python packages from builder
COPY --from=build-env /usr/local/bin/uv /usr/local/bin/uvx /usr/local/bin/
COPY --from=build-env /usr/local/lib/python3.13 /usr/local/lib/python3.13

# Copy OpenWhisk runtime and proxy binary
COPY --from=builder /go/bin/proxy /bin/proxy
ADD bin/compile /bin/compile
ADD lib/launcher.py /lib/launcher.py

# Prepare /action
WORKDIR /action
RUN chown nobody:root /action && chmod 0775 /action

USER nobody
ENTRYPOINT ["/bin/proxy"]
