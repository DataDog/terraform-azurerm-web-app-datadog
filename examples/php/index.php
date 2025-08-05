<?php
require_once 'vendor/autoload.php';

#[\DDTrace\Trace]
function helloWorld() {
    echo "Hello World from PHP!";
}

helloWorld();
