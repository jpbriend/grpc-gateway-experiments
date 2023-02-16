#!/usr/bin/env bash

MACHINE=$(uname -m)
case $MACHINE in
  'x86_64')
    curl -sSfL https://github.com/protocolbuffers/protobuf/releases/download/v$1/protoc-$1-osx-x86_64.zip > /tmp/protoc.zip
    ;;
  'arm64')
    curl -sSfL https://github.com/protocolbuffers/protobuf/releases/download/v$1/protoc-$1-osx-aarch_64.zip > /tmp/protoc.zip
    ;;
  *)
    echo "Unhandled POSIX Machine $MACHINE - refer to https://en.wikipedia.org/wiki/Uname"
    exit 99
    ;;
esac