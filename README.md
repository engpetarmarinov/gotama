# gotama - **go** **ta**sk **ma**nager
Simple scalable system that can schedule and execute tasks.

![gotama logo](./docs/assets/gotama-logo.png)

* Submit a new task for immediate or periodic execution.
* List all the submitted tasks and their current status.
* View a single task’s details and status.
* Tasks can be run once or recurring.

## Architecture
![architecture](./docs/assets/architecture.png)

## Why?
TODO

## Run locally
Prerequisites
* docker and docker-compose

Start the manager service, the broker (redis) and the workers
```bash
docker-compose up -d
```
or run natively with go
```bash
export $(grep -v '^#' local.env | xargs)
go run cmd/gotama-manager/main.go&
go run cmd/gotama-worker/main.go&
go run cmd/gotama-worker/main.go&
go run cmd/gotama-worker/main.go&
```
## RESTful API
Add a task
```bash
curl --location 'http://localhost:8080/api/v1/tasks' \
--header 'Content-Type: application/json' \
--data-raw '{
    "name": "email",
    "type": "recurring",
    "period": "5s",
    "payload": {
        "to": "eng.petar.marinov@gmail.com",
        "title": "Reminder",
        "body": "Take a break!"
    }
}'
```
Get a task
```bash
curl --location 'http://localhost:8080/api/v1/tasks/11ef259c-8523-42e4-8568-9d167dbba9da'
```
Get a list of tasks
```bash
curl --location 'http://localhost:8080/api/v1/tasks?limit=100&offset=0'
```
Update a task
```bash
curl --location --request PUT 'http://localhost:8080/api/v1/tasks/11ef259c-8523-42e4-8568-9d167dbba9da' \
--header 'Content-Type: application/json' \
--data-raw '{
    "name": "email",
    "type": "recurring",
    "period": "5s",
    "payload": {
        "to": "eng.petar.marinov@gmail.com",
        "title": "Reminder",
        "body": "Take a break!"
    }
}'
```
Delete a task
```bash
curl --location --request DELETE 'http://localhost:8080/api/v1/tasks/11ef259c-8523-42e4-8568-9d167dbba9da'
```