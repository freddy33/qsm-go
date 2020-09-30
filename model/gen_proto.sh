#!/usr/bin/env bash

protoc --proto_path=m3api/ --go_out=m3api/ --js_out=import_style=commonjs,binary:. m3api/*.proto

sed -i '' -E 's/\"([xyzd]|[a-z_]+_id|[a-z_]*dist|[a-z_]*time|growth_[a-z_]+),omitempty\"/\"\1\"/g;' m3api/*.pb.go
