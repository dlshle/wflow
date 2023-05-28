#!/bin/sh

protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/wflow.proto
cd admin-web
npm run protoc