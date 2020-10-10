#!/usr/bin/env bash

protoc --proto_path=m3api/ --go_out=m3api/ --js_out=import_style=commonjs,binary:. m3api/*.proto

sed -i '' -E '/( [u]*int32 | int64 | string ).*json:/ s/json:\"([a-z_]+)(,omitempty)*\"/json:\"\1\" query:\"\1\"/g;' m3api/*.pb.go
sed -i '' -E 's/json:\"-\"/json:\"-\" query:\"-\"/g;' m3api/*.pb.go
