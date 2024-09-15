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

# Use AdoptOpenJDK's JDK17
FROM ibm-semeru-runtimes:open-17.0.12.1_7-jdk-focal

RUN rm -rf /var/lib/apt/lists/* \
    && apt-get clean \
    && apt-get update \
    && apt-get -y --no-install-recommends upgrade \
    && apt-get -y --no-install-recommends install locales python vim \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* \
    && locale-gen en_US.UTF-8

ENV LANG="en_US.UTF-8" \
	LANGUAGE="en_US:en" \
	LC_ALL="en_US.UTF-8" \
	VERSION=17 \
	UPDATE=0.12 \
	BUILD=1_7

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
RUN chmod +x /bin/compile

# log initialization errors 
ENV OW_LOG_INIT_ERROR=1
# the compiler
ENV OW_COMPILER=/bin/compile
# default name to use when a jar is submitted as action
ENV OW_SAVE_JAR=exec.jar

ENTRYPOINT ["/bin/proxy"]
