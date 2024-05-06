/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

//--web true
//--docker ghcr.io/nuvolaris/runtime-nodejs-v21:3.1.0-mastrogpt.2403032235
//--param OPENAI_API_KEY $OPENAI_API_KEY
//--param OPENAI_API_HOST $OPENAI_API_HOST

const { OpenAIClient, AzureKeyCredential } = require("@azure/openai");

async function main(args) {
    const key = args.OPENAI_API_KEY || process.env.OPENAI_API_KEY
    const host = args.OPENAI_API_HOST || process.env.OPENAI_API_HOST
    const model = "gpt-35-turbo"
    const AI = new OpenAIClient(host, new AzureKeyCredential(key))
    const input = args.input || ""
    let answer = "Please provide an input parameter."
    if (input != "") {
        //const { id, created, choices, usage } =
        const request = [
            { role: "system", content: "You are a helpful assistant." },
            { role: "user", content: input },
        ]
        const response = await AI.getChatCompletions(model, request);
        answer = response.choices[0].message.content
    }
    return { body: answer }
}