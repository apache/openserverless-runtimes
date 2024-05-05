<?php
  require('vendor/autoload.php');

  function main(array $args): array {
    $loader = new \Twig\Loader\FilesystemLoader('templates/');
    $twig = new \Twig\Environment($loader,['cache'=>false]);
    $template = $twig->load('hello.twig');
    $name = $args["name"] || "world";
    return ['body' => $template->render(['name'=>$name])];
  }