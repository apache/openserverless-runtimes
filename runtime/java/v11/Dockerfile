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

# rely on nuvolaris common runtime image
ARG COMMON=missing:missing
FROM ${COMMON} AS builder

FROM ubuntu:22.04

# install Python
RUN apt-get update && apt-get -y upgrade &&\
    apt-get -y install apt-utils python3.10 python3.10-venv curl sudo telnet inetutils-ping zip unzip locales vim java-common\
	&& locale-gen en_US.UTF-8

ENV LANG="en_US.UTF-8" \
	LANGUAGE="en_US:en" \
	LC_ALL="en_US.UTF-8" \
	VERSION=11 \
	UPDATE=0 \
	BUILD=22

# Use Amazon corretto
RUN \
    VER=11 ;\
    ARCH=$(if arch | grep x86_64 >/dev/null ; then echo "x64" ; else echo "aarch64" ; fi) ;\
    curl -sL https://corretto.aws/downloads/latest/amazon-corretto-${VER}-${ARCH}-linux-jdk.deb >jdk.deb ;\
    dpkg -i jdk.deb ; rm jdk.deb 	

RUN locale-gen en_US.UTF-8 ;\
    mkdir -p /javaAction/action /usr/java/src /usr/java/lib

WORKDIR /javaAction

COPY --from=builder /go/bin/proxy /bin/proxy

ADD https://search.maven.org/remotecontent?filepath=com/google/code/gson/gson/2.9.0/gson-2.9.0.jar /usr/java/lib/gson-2.9.0.jar
ADD lib/src/Launcher.java /usr/java/src/Launcher.java
RUN cd /usr/java/src ;\
    javac -cp /usr/java/lib/gson-2.9.0.jar Launcher.java ;\
    jar cvf /usr/java/lib/launcher.jar *.class
ADD bin/compile /bin/compile
ENV OW_COMPILER=/bin/compile
ENV OW_SAVE_JAR=exec.jar
# using the runtime name to identify the execution environment
ENV OW_EXECUTION_ENV=openserverless-runtime-java/v11
ENTRYPOINT /bin/proxy
