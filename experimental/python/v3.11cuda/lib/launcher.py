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
from sys import stdin
from sys import stdout
from sys import stderr
from os import fdopen
import sys, os, json, traceback, warnings
import threading, collections
from pathlib import Path
import traceback

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
      sys.stderr.write("Invalid virtualenv. Zip file does not include 'activate_this.py'.\n")
      sys.exit(1)
except Exception:
  traceback.print_exc(file=sys.stderr, limit=0)
  sys.exit(1)

# now import the action as process input/output
import main__

out = fdopen(3, "wb")
if os.getenv("__OW_WAIT_FOR_ACK", "") != "":
    out.write(json.dumps({"ok": True}, ensure_ascii=False).encode('utf-8'))
    out.write(b'\n')
    out.flush()
    

env = os.environ

SETUP="_setup"
hash = env.get("__OW_CODE_HASH")
if hash: 
  SETUP="/tmp/"+hash
SETUP_DONE=SETUP+"_done"

# lanched as a thread to execute the setup
def setup_thread(setup, payload):
  with open(SETUP, 'w', buffering=1) as file:
    file.write("Setup thread started.\n")
    try:
      setup(payload, file)
      Path(SETUP_DONE).touch(exist_ok=True)
      file.write("Setup thread ended successfully.\n")
    except:
      traceback.print_exc(file=file)
 
while True:
  line = stdin.readline()
  if not line: break
  args = json.loads(line)
  
  payload = {}
  for key in args:
    if key == "value":
      payload = args["value"]
    else:
      env["__OW_%s" % key.upper()]= args[key]

  # if there is a setup
  if hasattr(main__, 'setup'):
    # if setup is not complete
    if not os.path.exists(SETUP_DONE):
      # if setup is running
      if os.path.exists(SETUP):
         payload['setup_status'] = Path(SETUP).read_text()
      else:
        payload['setup_status'] = "Setup thread started.\n"
        thread = threading.Thread(target=setup_thread, args=(main__.setup, payload,))
        thread.start()
        print("started setup", file=sys.stderr)

  res = {}
  try:
    res = main__.main(payload)
  except Exception as ex:
    print(traceback.format_exc(), file=stderr)
    res = {"error": str(ex)}
  out.write(json.dumps(res, ensure_ascii=False).encode('utf-8'))
  out.write(b'\n')
  stdout.flush()
  stderr.flush()
  out.flush()
