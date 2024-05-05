<?php
//--web true
//--kind php:default
function main(array $args): array
{
  $name = array_key_exists("name", $args) ? $args["name"] : "world";
  return ['body' => sprintf('Hello %s', $name)];
}