<?php
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