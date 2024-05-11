<?php
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
//--kind php:default
//--param OPENAI_API_KEY $OPENAI_API_KEY
//--param OPENAI_API_HOST $OPENAI_API_HOST
function main(array $args): array
{
  $openaiKey = array_key_exists('OPENAI_API_KEY',$args) ? $args['OPENAI_API_KEY'] : null;
  $openaiHost = array_key_exists('OPENAI_API_HOST',$args) ? $args['OPENAI_API_HOST'] : null;
  
  $model = "gpt-35-turbo";

  if (empty($openaiHost)) {
    $openaiHost = 'openai.nuvolaris.io';
  }
  if (empty($openaiKey)) {
    return ['error'=>"OpenAI Key is not set"];
  }

  print "Sending request to $openaiHost";

  $client = OpenAI::factory()
  ->withApiKey($openaiKey)
  ->withBaseUri($openaiHost)
  ->make();

  $response = $client->models()->list();

  /*$response = $client->chat()->create([
    'model' => $model,
    'messages' => [
        ['role' => 'user', 'content' => 'Hello!'],
    ],
  ]);*/

  return ['body' => print_r($response, true)];
}