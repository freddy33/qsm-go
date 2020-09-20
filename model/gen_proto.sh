#!/usr/bin/env bash

protoc --proto_path=m3api/ --go_out=m3api/ --js_out=import_style=commonjs,binary:. m3api/*.proto
