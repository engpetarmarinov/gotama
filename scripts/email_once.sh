#!/bin/bash

for ((i=1; i<=10000; i++)); do
    curl --location 'http://localhost:8080/api/v1/tasks' \
    --header 'Content-Type: application/json' \
    --data-raw '{
        "name": "email",
        "type": "once",
        "payload": {
            "to": "gotama@gotama.io",
            "title": "Reminder '"$i"'",
            "body": "Take a break!"
        }
    }'
done