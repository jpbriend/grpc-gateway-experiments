#!/usr/bin/env bash

PROTOC=$1
PROTO_PATH=$2
PROTOC_GEN_GRPC_GATEWAY=$3
GATEWAY_OUT=$4
PROTO_SOURCE=$5

# Search for all _gateway.yaml files
for i in $(find $PROTO_SOURCE -name '*_gateway.yaml');
do
  GATEWAY_FILE=$i
  PROTO_FILE=$(echo $GATEWAY_FILE | sed 's/_gateway.yaml/.proto/g')
  # Search if the _gateway.yaml has a corresponding .proto file (it should)
  if [ -f $PROTO_FILE ]; then
    echo "➡️Generating gateway for $GATEWAY_FILE"
    $PROTOC \
			--plugin protoc-gen-grpc-gateway=$PROTOC_GEN_GRPC_GATEWAY \
			--grpc-gateway_out="$GATEWAY_OUT" \
			--grpc-gateway_opt=paths=source_relative \
			--grpc-gateway_opt=generate_unbound_methods=false \
			--grpc-gateway_opt=logtostderr=true \
			--grpc-gateway_opt=grpc_api_configuration=$GATEWAY_FILE \
  		-I=$PROTO_PATH \
			$PROTO_FILE
  fi;
done;