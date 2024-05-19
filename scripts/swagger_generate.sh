#!/bin/bash

swagger generate spec -o ./docs/swagger.yaml --scan-models
swagger validate ./docs/swagger.yaml