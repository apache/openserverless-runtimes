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

FROM php:8.5-cli-trixie
ADD --chmod=0755 https://github.com/mlocati/docker-php-extension-installer/releases/latest/download/install-php-extensions /usr/local/bin/

# install PHP extensions
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    apt-get -y update \
    # Upgrade installed packages to get latest security fixes if the base image does not contain them already.
    && apt-get upgrade -y --no-install-recommends \
    && apt-get -y install --no-install-recommends \
      unzip \
      libfreetype6 \
      libicu76 \
      libjpeg62-turbo \
      libpng16-16t64 \
      libssl3t64 \
      libxml2 \
      libzip5 \
      libpq5 \
      jq \
      zip \
      libfreetype6-dev \
      libicu-dev \
      libjpeg-dev \
      libpng-dev \
      libssl-dev \
      libxml2-dev \
      libzip-dev \
      postgresql-server-dev-17 \
    && install-php-extensions  \
      bcmath \
      gd \
      intl \
      mysqli \
      mongodb \
      opcache \
      pgsql \
      pdo_mysql \
      pdo_pgsql \
      redis \
      soap \
      zip \
    && apt-get purge -y --auto-remove \
      $PHPIZE_DEPS \
      libfreetype6-dev \
      libicu-dev \
      libjpeg-dev \
      libpng-dev \
      libssl-dev \
      libxml2-dev \
      libzip-dev \
      postgresql-server-dev-17 \
    && apt-get autoremove -y \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* /usr/src/php \
    && mkdir -p /tmp/opcache && chown nobody:root /tmp/opcache && chmod 0775 /tmp/opcache

COPY php.ini /usr/local/etc/php

# install composer
RUN curl -s -f -L -o /tmp/installer.php https://getcomposer.org/installer \
    && php /tmp/installer.php --no-ansi --install-dir=/usr/bin --filename=composer \
    && rm /tmp/installer.php \
    && composer --ansi --version --no-interaction --no-plugins --no-scripts

# install default Composer dependencies
WORKDIR /app
RUN mkdir -p /app/.composer /app/action
COPY composer.json /app/.composer/
RUN --mount=type=cache,target=/tmp/composer-cache,uid=0,gid=0 \
    cd /app/.composer \
    && COMPOSER_CACHE_DIR=/tmp/composer-cache /usr/bin/composer install --no-plugins --no-scripts --prefer-dist --no-dev -o --classmap-authoritative \
    && rm -f composer.lock \
    && chown -R nobody:root /app \
    && chmod -R 0775 /app

# install proxy binary along with compile and launcher scripts
COPY --from=builder /go/bin/proxy /bin/proxy
COPY --chmod=0755 compile.php /bin/compile.php
COPY --chmod=0755 extend /bin/extend
COPY --chmod=0755 runner.php /bin/runner.php
ENV OW_COMPILER=/bin/compile.php
ENV OW_EXECUTION_ENV=apacheopenserverless/action-php-v8.5
ENV COMPOSER_HOME=/app/.composer
ENV HOME=/app

EXPOSE 8080
USER nobody
ENTRYPOINT [ "/bin/proxy" ]