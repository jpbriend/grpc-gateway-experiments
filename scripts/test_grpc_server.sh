#!/usr/bin/env bash
grpcurl --plaintext --proto ./proto/potato/potato.proto localhost:8080 example.potato.PotatoService/GetPotatoes