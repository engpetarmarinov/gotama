#!/bin/bash

for ((i=1; i<=100; i++)); do
    curl --location 'http://localhost:8080/api/v1/tasks' \
    --header 'Content-Type: application/json' \
    --data-raw '{
        "name": "email",
        "type": "recurring",
        "period": "5s",
        "payload": {
            "to": "gotama@gotama.io",
            "title": "Recurring Reminder '"$i"'",
            "body": "Take a break!"
        }
    }'
done