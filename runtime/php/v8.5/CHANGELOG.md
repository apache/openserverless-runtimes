<!--
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
-->

## 2025-12-26
- **Upgraded to PHP 8.5**
  - Updated base image to `php:8.5-cli-trixie`
  - Upgraded to Debian Trixie for latest system libraries and security updates

- **Updated System Dependencies**
  - Upgraded ICU library from libicu72 to libicu76 for improved internationalization support
  - Upgraded libzip from version 4 to 5 for better compression support
  - Upgraded OpenSSL libraries to libssl3t64 (transition package) for enhanced security
  - Upgraded libpng16-16 to libpng16-16t64 (transition package)
  - Upgraded PostgreSQL development headers from version 15 to 17

- **Performance Optimizations**
  - **Enabled JIT (Just-In-Time) compilation** with tracing mode for 2-3x performance improvement on CPU-intensive workloads
    - `opcache.jit=tracing` - optimal for web applications
    - `opcache.jit_buffer_size=128M` - dedicated memory for JIT compiled code
    - Enabled hot function/loop/return/side_exit optimizations
  - **Enhanced OPcache configuration**
    - Increased `max_accelerated_files` to 16,229 (prime number for better hash distribution)
    - Added `memory_consumption=256MB` for opcache
    - Added `interned_strings_buffer=16MB` to reduce memory usage for duplicate strings
    - Enabled file cache at `/tmp/opcache` for faster startup times
    - Disabled save_comments and file_cache_consistency_checks for better performance
  - **Modernized runner.php with PHP 8.4+ features**
    - Replaced if/elseif chains with `match` expressions for better performance
    - Used bitwise operators for error type checking (faster than `in_array()`)
    - Implemented strict JSON handling with `JSON_THROW_ON_ERROR` and `JSON_BIGINT_AS_STRING`
    - Pre-opened file handles with disabled buffering for faster I/O
    - Added `get_debug_type()` for improved type information
    - Pre-allocated error message constants to reduce string allocations
    - Added early function validation before processing

- **Docker Build Optimizations**
  - Added BuildKit cache mounts for APT and Composer caches to speed up rebuilds
  - Added `--classmap-authoritative` flag for optimized autoloader generation
  - Pre-created opcache directory with correct permissions
  - Optimized Dockerfile for reduced image size
  - Added APT cache cleanup to remove package lists
  - Added Composer cache cleanup after dependency installation
  - Removed Composer installer after installation
  - Consolidated file permission settings using COPY --chmod
  - Added security improvements: EXPOSE 8080 and USER nobody

## Apache 1.20.0
- Initial Release

- Added: PHP: 8.5
- Added: PHP extensions in addition to the standard ones:
    - bcmath
    - curl
    - gd
    - intl
    - mbstring
    - mysqli
    - pdo_mysql
    - pdo_pgsql
    - pdo_sqlite
    - soap
    - zip
    - mongo
- Added: Composer packages:
    - [guzzlehttp/guzzle](https://packagist.org/packages/guzzlehttp/guzzle): 7.8.1
    - [ramsey/uuid](https://packagist.org/packages/ramsey/uuid): 4.7.5