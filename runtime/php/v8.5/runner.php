<?php
/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// open fd/3 as that's where we send the result
$fd3 = fopen('php://fd/3', 'wb');
if ($fd3 === false) {
    fwrite(STDERR, "Failed to open file descriptor 3\n");
    exit(1);
}

// Set stream to binary mode and disable buffering for performance
stream_set_write_buffer($fd3, 0);

// Pre-open stderr and stdout for better performance
$stderr = fopen('php://stderr', 'wb');
$stdout = fopen('php://stdout', 'wb');
stream_set_write_buffer($stderr, 0);
stream_set_write_buffer($stdout, 0);

// Register a shutdown function so that we can fail gracefully when a fatal error occurs
register_shutdown_function(static function () use ($fd3, $stderr) {
    $error = error_get_last();
    if ($error && ($error['type'] & (E_ERROR | E_CORE_ERROR | E_COMPILE_ERROR | E_USER_ERROR))) {
        fwrite($stderr, "An error occurred running the action.\n");
        fwrite($fd3, "An error occurred running the action.\n");
    }
    fclose($fd3);
    fclose($stderr);
});

require 'vendor/autoload.php';
require 'index.php';

// retrieve main function
$__functionName = $argv[1] ?? 'main';

// Validate function exists
if (!function_exists($__functionName)) {
    fwrite($stderr, "Function '{$__functionName}' not found\n");
    exit(1);
}

// Pre-allocate constants
const ERROR_MSG = 'An error occurred running the action.';
const DICT_ERROR_MSG = 'The action did not return a dictionary or array.';

// read stdin
while (($line = fgets(STDIN)) !== false) {
    // call the function - optimized JSON decoding with flags
    $data = json_decode($line, true, 512, JSON_THROW_ON_ERROR | JSON_BIGINT_AS_STRING) ?? [];

    // convert all parameters other than value to environment variables
    foreach ($data as $key => $value) {
        if ($key !== 'value') {
            $envKeyName = '__OW_' . strtoupper($key);
            $_ENV[$envKeyName] = $value;
            putenv($envKeyName . '=' . $value);
        }
    }

    $values = $data['value'] ?? [];
    try {
        $result = $__functionName($values);

        // convert result to an array if we can - optimized with match expression
        $result = match(true) {
            $result === null => [],
            is_object($result) && method_exists($result, 'getArrayCopy') => $result->getArrayCopy(),
            $result instanceof stdClass => (array)$result,
            default => $result
        };

        // process the result
        if (!is_array($result)) {
            $errorMsg = 'Result must be an array but has type "' . get_debug_type($result) . '": ' . $result;
            fwrite($stderr, $errorMsg);
            fwrite($stdout, DICT_ERROR_MSG);
            $result = (string)$result;
        } else {
            // cast result to an object for json_encode to ensure that an empty array becomes "{}
            $result = json_encode((object)$result, JSON_THROW_ON_ERROR | JSON_UNESCAPED_SLASHES | JSON_UNESCAPED_UNICODE);
        }
    } catch (JsonException $e) {
        fwrite($stderr, "JSON Error: {$e->getMessage()}\n");
        $result = ERROR_MSG;
    } catch (Throwable $e) {
        fwrite($stderr, (string)$e);
        $result = ERROR_MSG;
    }

    // ensure that the sentinels will be on their own lines
    fwrite($stderr, "\n");
    fwrite($stdout, "\n");

    // send result to fd/3
    fwrite($fd3, $result . "\n");
}