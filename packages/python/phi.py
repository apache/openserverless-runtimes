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

from subprocess import run

def login(args, status):
    from huggingface_hub import login, whoami
    try:
        whoami()
        status.write("already logged in\n")
        return True
    except:
       try:
          login(token=args.get("hf_token", ""))
          status.write("logged in\n")
          return True
       except:
          status.write("cannot log in - did you provide a correct hf_token?\n")
          return False

def setup(args, status):
    from subprocess import run
    run(["pip", "install", "transformers", "--upgrade"])
    status.write("downloading model\n")
    from transformers import pipeline
    pipeline(model="microsoft/phi-1_5")
    run(["pip", "install", "huggingface_hub"])
    if login(args, status):
        status.write("logged in\n")
    
    
chat = None
def main(args):
    global chat
    if "setup_status" in args:
        return {"body": args['setup_status']}
    
    from transformers import pipeline
    if not chat:
        chat = pipeline(model="microsoft/phi-1_5")
    
    return {
        "body": chat(args.get("input", "who are you"), max_new_tokens=50)
    }
