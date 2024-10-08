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
    
from subprocess import run
import os

def setup(status):
    status.append("installing huggingface_hub")
    run(["pip", "install", "huggingface_hub"])
    status.append("installing protobuf")
    run(["pip", "install", "protobuf"])
    status.append("installing sentencepiece")
    run(["pip", "install", "sentencepiece"])
    status.append("downloading mistral model - 14GB be patient!")
    from transformers import pipeline
    chatbot = pipeline("text-generation", model="mistralai/Mistral-7B-Instruct-v0.3")

def login(args):
    from huggingface_hub import login, whoami
    try:
        whoami()
        print("already logged in")
        return True
    except:
       try:
          login(token=args.get("hf_token", ""))
          print("logged in")
          return True
       except:
          return False

def main(args):
    if "setup_status" in args:
        res = "\n".join(args['setup_status'])
        return { "body": res }
 
    if not login(args):
        return {"body": "cannot login - is the hf_token provided correctly"}   
   
    return {
        "body": "ok"
    }

