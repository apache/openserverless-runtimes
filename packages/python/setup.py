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

#--web true
#--kind python:default

def setup(args, status):
    import sys, time
    print(args)
    n = int(args.get("count", "10"))
    for i in range(1,n):
        print(f"setup {i}", file=sys.stderr)
        status.append(f"setup level {i}")
        time.sleep(5)

def main(args):
    if "setup_status" in args:
        return { "body":  args['setup_status'] }
    name = args.get("name", "world")
    return {
        "body": f"Hello, {name}."
    }

