#!/bin/bash

# foo is designed to fail and test fail-retry mechanism
curl --location 'http://localhost:8080/api/v1/tasks' \
    --header 'Content-Type: application/json' \
    --data-raw '{
        "name": "foo",
        "type": "once",
        "payload": {
            "bar "bar",
            "baz": "baz"
        }
    }'
