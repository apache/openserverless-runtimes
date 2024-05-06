<!--
  ~ Licensed to the Apache Software Foundation (ASF) under one
  ~ or more contributor license agreements.  See the NOTICE file
  ~ distributed with this work for additional information
  ~ regarding copyright ownership.  The ASF licenses this file
  ~ to you under the Apache License, Version 2.0 (the
  ~ "License"); you may not use this file except in compliance
  ~ with the License.  You may obtain a copy of the License at
  ~
  ~   http://www.apache.org/licenses/LICENSE-2.0
  ~
  ~ Unless required by applicable law or agreed to in writing,
  ~ software distributed under the License is distributed on an
  ~ "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
  ~ KIND, either express or implied.  See the License for the
  ~ specific language governing permissions and limitations
  ~ under the License.
-->

export NUV_ROOT=/workspaces/all/olaris
cd /workspaces/all/nuvolaris-runtimes-ng

nuv ide python

## start
nuv ide python clean DIR=$PWD/runtime/python/test/withreqs
ls -l $PWD/runtime/python/test/withreqs.zip

# not found
nuv ide python environment DIR=$PWD/runtime/python/test/withreqs
# built
nuv ide python environment DIR=$PWD/runtime/python/test/withreqs
# not rebuilt

echo "" >>$PWD/runtime/python/test/withreqs/requirements.txt
nuv ide python environment DIR=$PWD/runtime/python/test/withreqs
nuv ide python environment DIR=$PWD/runtime/python/test/withreqs
# modified and rebuilt

nuv ide python compile DIR=$PWD/runtime/python/test/withreqs

unzip -l $PWD/runtime/python/test/withreqs.zip
## end

## start
nuv ide python clean DIR=$PWD/runtime/python/test/multifile
ls -l $PWD/runtime/python/test/multifile.zip

# not found
nuv ide python environment DIR=$PWD/runtime/python/test/multifile
# built
nuv ide python environment DIR=$PWD/runtime/python/test/multifile
# not rebuilt

# modified and rebuilt
nuv ide python compile DIR=$PWD/runtime/python/test/multifile
unzip -l $PWD/runtime/python/test/multifile.zip
## end

## start
nuv ide python clean DIR=$PWD/runtime/python/test/multifile
ls -l $PWD/runtime/python/test/multifile.zip

# not found
nuv ide python environment DIR=$PWD/runtime/python/test/multifile
# built
nuv ide python environment DIR=$PWD/runtime/python/test/withreqs
# not rebuilt

echo "" >>$PWD/runtime/python/test/withreqs/requirements.txt
nuv ide python environment DIR=$PWD/runtime/python/test/withreqs
nuv ide python environment DIR=$PWD/runtime/python/test/withreqs
# modified and rebuilt

nuv ide python compile DIR=$PWD/runtime/python/test/withreqs

unzip -l $PWD/runtime/python/test/withreqs.zip
## end





