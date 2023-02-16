#!/usr/bin/env bash

curl -sSfL https://github.com/protocolbuffers/protobuf/releases/download/v$1/protoc-$1-linux-$(uname -m).zip > /tmp/protoc.zip