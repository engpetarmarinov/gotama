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

Start the manager service, the queue (redis) and the workers
```bash
docker-compose up -d
```
or run natively
```bash
export $(grep -v '^#' .env | xargs)
go run cmd/gotama-manager/main.go
```
TODO