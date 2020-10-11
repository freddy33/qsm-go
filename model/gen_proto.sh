#!/usr/bin/env bash

protoc --proto_path=m3api/ --go_out=m3api/ --js_out=import_style=commonjs,binary:. m3api/*.proto

# Remove omitempty on simple types and add query metadata
sed -i '' -E '/( [u]*int32 | int64 | string ).*json:/ s/json:\"([a-z_]+)(,omitempty)*\"/json:\"\1\" query:\"\1\"/g;' m3api/*.pb.go
# Keep omitempty on array and complex types and add query metadata
sed -i '' -E '/( \*PointMsg | \[\]int32 | \[\]int64 ).*json:/ s/json:\"([a-z_]+)(,omitempty)*\"/json:\"\1\2\" query:\"\1\"/g;' m3api/*.pb.go
# Make sure all other types are not sent via query params
sed -i '' -E '/( \*PointMsg | \[\]int32 | \[\]int64 | [u]*int32 | int64 | string )/! s/json:\"([a-z_]+)(,omitempty)*\"/json:\"\1\2\" query:\"-\"/g;' m3api/*.pb.go
sed -i '' -E 's/json:\"-\"/json:\"-\" query:\"-\"/g;' m3api/*.pb.go
