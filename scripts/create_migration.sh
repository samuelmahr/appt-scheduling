#!/usr/bin/env bash
migrate create -ext sql -dir ./migrations -format unix "$1"