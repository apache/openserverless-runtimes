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

# Use AdoptOpenJDK's JDK25
FROM ibm-semeru-runtimes:open-jdk-25.0.4.0-jdk-noble

RUN rm -rf /var/lib/apt/lists/* \
    && apt-get clean \
    && apt-get update \
    && apt-get -y --no-install-recommends upgrade \
    && apt-get -y --no-install-recommends install locales python3 maven\
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* \
    && locale-gen en_US.UTF-8 \
    && mkdir -p /javaAction/action /usr/java/src /usr/java/lib

ENV LANG="en_US.UTF-8" \
	LANGUAGE="en_US:en" \
	LC_ALL="en_US.UTF-8" \
	VERSION=25 \
	UPDATE=0.4 \
	BUILD=0

WORKDIR /javaAction

COPY --from=builder /go/bin/proxy /bin/proxy
COPY lib/src/Launcher.java /usr/java/src/Launcher.java
COPY bin/compile bin/extend /bin/
COPY pom.xml /tmp

RUN chmod +x /bin/compile /bin/extend \
    && mkdir -p /javaAction/.m2/repository \
    && chown -R nobody:nogroup /javaAction \
    && chmod -R u+rwX /javaAction \
    && extend \
    && cd /usr/java/src \
    && GSON_JAR=$(find /javaAction/.m2/repository -name "gson-*.jar") \
    && echo "Using GSON_JAR: $GSON_JAR" \
    && javac -cp "$GSON_JAR" Launcher.java \
    && jar cvf /usr/java/lib/launcher.jar *.class \
    && chmod 644 /usr/java/lib/launcher.jar \
    && rm -f /usr/java/src/*.class
    

# log initialization errors
ENV OW_LOG_INIT_ERROR=1
# the compiler
ENV OW_COMPILER=/bin/compile
# default name to use when a jar is submitted as action
ENV OW_SAVE_JAR=exec.jar
ENV MAVEN_CONFIG=/javaAction/.m2
ENV MAVEN_OPTS="-Dmaven.repo.local=/javaAction/.m2/repository"

EXPOSE 8080
USER nobody

ENTRYPOINT ["/bin/proxy"]
