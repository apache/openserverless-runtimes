#--web true
#--docker https://registry.hub.docker.com/94lama/python:flask

import json
from flask import Flask, request
from hello import hello

app = Flask(__name__)

@app.route("/")
def home():
    return "Hello, app!"

@app.route("/post", methods=["POST"])
def post():
    if request.is_json:
        return request.json
    elif request.form:
        return dict(request.form)
    else:
        return request.data.decode('utf-8')

@app.route("/put", methods=["PUT"])
def put():
    if request.is_json:
        return request.json
    elif request.form:
        return dict(request.form)
    else:
        return request.data.decode('utf-8')

@app.route("/delete", methods=["DELETE"])
def delete():
    if request.is_json:
        return request.json
    elif request.form:
        return dict(request.form)
    else:
        return request.data.decode('utf-8')

@app.route("/test")
def test_param_query():
    return json.dumps(request.args) # This should return a JSON
