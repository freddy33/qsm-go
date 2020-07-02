#!/usr/bin/env bash

protoc --proto_path=m3api/ --go_out=m3api/ m3api/*.proto

