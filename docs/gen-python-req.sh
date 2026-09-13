# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.
# default packages available across all runtimes


# this is not a script but a sequence of commands to be run manually

DIR=sys
DIR=v3.11
DIR=v3.12
DIR=v3.13
cd "$(dirname $0)"
BASE=$(pwd)
for DIR in sys v3.11 v3.12 v3.13
do
    # pick your dir
    cd $BASE/../runtime/python/$DIR

docker run -i -v $PWD:/mnt $(awk '/FROM python/{ print $2}' Dockerfile) bash <<EOF
cd /mnt
rm requirements.txt

# pip
pip install pip-tools
pip-compile requirements.in
{ head -n 16 requirements.in; cat requirements.txt; } > requirements.new && mv requirements.new requirements.txt
EOF

done

# uv
#head -n 16 requirements.in >requirements.txt
#uv pip compile requirements.in  >>requirements.txt

