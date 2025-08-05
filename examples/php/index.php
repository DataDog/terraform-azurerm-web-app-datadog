<?php
require __DIR__ . '/vendor/autoload.php';

#[\DDTrace\Trace]
function helloWorld() {
    echo "Hello World from PHP!";
}

helloWorld();
