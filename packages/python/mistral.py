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
import os, io

def login(args, status: io.TextIOWrapper):
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

def setup(args, status: io.TextIOWrapper):
    #status.write("installing huggingface_hub")
    #run(["pip", "install", "huggingface_hub"])
    #status.write("installing accelerate")  
    #run(["pip", "install", "accelerate"])
    #status.write("installing protobuf")  
    #run(["pip", "install", "protobuf"])
    #status.write("installing sentencepiece")
    #run(["pip", "install", "sentencepiece"])
    #status.write("installing mistral_inference")
    #run(["pip", "install", "mistral_inference"])
    if login(args, status):
        status.write("downloading mistral model - it is 14GB be patient!\n")
        from transformers import pipeline
        pipeline("text-generation", model="mistralai/Mistral-7B-Instruct-v0.3")
        status.write("mistral model - downloaded\n")

def main(args):
    
    print(repr(args))

    if "setup_status" in args:
        res = str(args['setup_status']).split("\n")
        print(repr(res))
        return { "body": "setup in progress", "statusCode": 202 }
    
    from huggingface_hub import  whoami 
    return {
        "body": whoami()
    }

    
