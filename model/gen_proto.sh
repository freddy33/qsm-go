#!/usr/bin/env bash

protoc --proto_path=m3api/ --go_out=m3api/ --js_out=library=m3api_js_libs,binary:. m3api/*.proto
