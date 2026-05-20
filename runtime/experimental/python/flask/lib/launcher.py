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
from __future__ import print_function
from sys import stdin, stdout, stderr
from os import fdopen
from io import BytesIO
import base64
import urllib.parse
import sys, os, json, traceback, requests


try:
  # if the directory 'virtualenv' is extracted out of a zip file
  path_to_virtualenv = os.path.abspath('./virtualenv')
  if os.path.isdir(path_to_virtualenv):
    # activate the virtualenv using activate_this.py contained in the virtualenv
    activate_this_file = path_to_virtualenv + '/bin/activate_this.py'
    if not os.path.exists(activate_this_file): # try windows path
      activate_this_file = path_to_virtualenv + '/Scripts/activate_this.py'
    if os.path.exists(activate_this_file):
      with open(activate_this_file) as f:
        code = compile(f.read(), activate_this_file, 'exec')
        exec(code, dict(__file__=activate_this_file))
    else:
      stderr.write("Invalid virtualenv. Zip file does not include 'activate_this.py'.\n")
      sys.exit(1)
except Exception:
  traceback.print_exc(file=sys.stderr, limit=0)
  sys.exit(1)

# now import the action as process input/output
from main__ import app as app

out = fdopen(3, "wb")
if os.getenv("__OW_WAIT_FOR_ACK", "") != "":
    out.write(json.dumps({"ok": True}, ensure_ascii=False).encode('utf-8'))
    out.write(b'\n')
    out.flush()

env = os.environ

global response_status
global response_headers
response_status = None
response_headers = None

def start_response(status, headers):
  global response_status
  global response_headers
  response_status = status
  response_headers = headers
  return []

environ = {
  "wsgi.version": (1, 0),
  "wsgi.url_scheme": "HTTP",
  "wsgi.input": stdin,
  "wsgi.output": stdout,
  "wsgi.errors": stderr,
  "wsgi.multithread": False,
  "wsgi.multiprocess": True,
  "wsgi.run_once": True,
  "ACCEPT": "*/*",
  "PATH_INFO": "/",
  "CONTENT_TYPE": "application/json",
  "CONTENT_LENGTH": "0",
  "QUERY_STRING": "",
  "REQUEST_METHOD": "GET",
  "SERVER_PROTOCOL": "HTTP/1.1",
}

def reset_environ():
  environ["wsgi.input"]=stdin
  environ["ACCEPT"]="*/*"
  environ["CONTENT_TYPE"]="application/json"
  environ["CONTENT_LENGTH"] = "0"
  environ["PATH_INFO"] = "/"
  environ["QUERY_STRING"] = ""
  environ["REQUEST_METHOD"]="GET"

# Collect the response body
def build_response():
  response_body = b''
  response = app(environ, start_response)
  try:
    for chunk in response:
      if isinstance(chunk, str):
        chunk = chunk.encode('utf-8')
      response_body += chunk
  except Exception as e:
    stderr.write(f"Error building response: {e}\n")
    response_body = f'Error building response: {e}'.encode('utf-8')

  # Convert WSGI headers list to a dictionary
  headers_dict = {}
  if response_headers:
    for header_name, header_value in response_headers:
      headers_dict[header_name] = header_value

  # Parse status code from status string (e.g., "200 OK" -> 200)
  status_code = 200
  if response_status:
    try:
      status_code = int(response_status.split()[0])
    except (ValueError, IndexError):
      status_code = 200

  return {
    "statusCode": status_code,
    "headers": headers_dict,
    "body": response_body.decode('utf-8') if isinstance(response_body, bytes) else response_body,
  }

while True:
  line = stdin.readline()
  if not line: break
  args = json.loads(line)

  reset_environ()
  res = {}

  try:
    stderr.write("="*80 + "\n")
    stderr.write("DEBUG: NEW REQUEST\n")
    stderr.write("="*80 + "\n")
    stderr.write(f"DEBUG: Incoming args: {json.dumps(args, indent=2)}\n")
    stderr.write("-"*80 + "\n")
    
    # Initialize collections
    other_params = {}
    
    # Parse the input arguments to build the WSGI environ
    if "value" in args and isinstance(args["value"], dict):
      stderr.write("DEBUG: Processing value fields...\n")
      for k, v in args["value"].items():
        if k == "PREFERRED_URL_SCHEME": 
          environ["wsgi.url_scheme"] = v

        elif k == "API_URL": 
          environ["API_URL"] = v

        elif k == "__ow_method": 
          environ["REQUEST_METHOD"] = v.upper()
          stderr.write(f"DEBUG: ✓ Method set to: {v.upper()}\n")

        elif k == "__ow_path": 
          environ["PATH_INFO"] = v
          stderr.write(f"DEBUG: ✓ Path set to: {v}\n")

        elif k == "__ow_headers":
          if isinstance(v, dict):
            stderr.write(f"DEBUG: Processing headers: {list(v.keys())}\n")
            for k, v in v.items():
              if k.lower() == "x-scheme": environ["wsgi.url_scheme"] = v
              else: environ[k.upper().replace("-", "_")] = v

        elif not k.startswith("__ow_") and k not in ["action_name", "action_version", "activation_id", "deadline", "namespace", "transaction_id"]:
          # Collect all other params (these are the actual request data)
          other_params[k] = v
          stderr.write(f"DEBUG: ✓ Collected param: {k} = {v}\n")

    stderr.write("-"*80 + "\n")
    stderr.write(f"DEBUG: Total other_params collected: {len(other_params)} items\n")
    stderr.write(f"DEBUG: other_params = {json.dumps(other_params, indent=2)}\n")
    stderr.write("-"*80 + "\n")

    # Build body or query string based on method
    ct = environ.get("CONTENT_TYPE", "").lower()
    body_bytes = b""
    method = environ.get("REQUEST_METHOD", "GET").upper()

    stderr.write(f"DEBUG: REQUEST_METHOD = {method}\n")
    stderr.write(f"DEBUG: CONTENT_TYPE = {ct}\n")
    stderr.write(f"DEBUG: PATH_INFO = {environ.get('PATH_INFO')}\n")
    stderr.write("-"*80 + "\n")

    if method in ["POST", "PUT", "PATCH", "DELETE"]:
      stderr.write(f"DEBUG: Processing {method} request body...\n")
      # For POST/PUT/PATCH/DELETE, other_params go into the request body
      if other_params:
        if "json" in ct or not ct or ct == "application/json":
          payload = json.dumps(other_params)
          body_bytes = payload.encode("utf-8")
          stderr.write(f"DEBUG: ✓ Created JSON payload ({len(body_bytes)} bytes): {payload}\n")
        elif "form-urlencoded" in ct:
          payload = urllib.parse.urlencode(other_params, doseq=True)
          body_bytes = payload.encode("utf-8")
          stderr.write(f"DEBUG: ✓ Created form payload ({len(body_bytes)} bytes): {payload}\n")
        else:
          payload = json.dumps(other_params)
          body_bytes = payload.encode("utf-8")
          stderr.write(f"DEBUG: ✓ Created default JSON payload ({len(body_bytes)} bytes): {payload}\n")
      else:
        stderr.write(f"DEBUG: ⚠ No other_params for {method} request - body will be empty\n")
      
      environ["CONTENT_LENGTH"] = str(len(body_bytes))
      environ["wsgi.input"] = BytesIO(body_bytes)
      stderr.write(f"DEBUG: Set CONTENT_LENGTH = {len(body_bytes)}\n")
      stderr.write(f"DEBUG: Set wsgi.input to BytesIO with {len(body_bytes)} bytes\n")
    else:
      stderr.write(f"DEBUG: Processing GET request...\n")
      # For GET, other_params go into query string
      if other_params:
        environ["QUERY_STRING"] = urllib.parse.urlencode(other_params, doseq=True)
        stderr.write(f"DEBUG: ✓ Query string: {environ['QUERY_STRING']}\n")
      else:
        stderr.write(f"DEBUG: No query parameters\n")
      environ["wsgi.input"] = BytesIO(b"")

    stderr.write("-"*80 + "\n")
    stderr.write("DEBUG: Calling Flask app...\n")
    res = build_response()
    stderr.write("-"*80 + "\n")
    stderr.write(f"DEBUG: Response received:\n")
    stderr.write(f"  - statusCode: {res.get('statusCode')}\n")
    stderr.write(f"  - headers: {res.get('headers')}\n")
    stderr.write(f"  - body length: {len(res.get('body', ''))}\n")
    stderr.write(f"  - body preview: {res.get('body', '')[:200]}\n")
    stderr.write("="*80 + "\n")
    
  except Exception as e:
    stderr.write(f"Exception occurred: {str(e)}\n")
    traceback.print_exc(file=stderr)
    res = {"error": str(e)}

  out.write(json.dumps(res, ensure_ascii=False).encode('utf-8'))
  out.write(b'\n')
  stdout.flush()
  stderr.flush()
  out.flush()
