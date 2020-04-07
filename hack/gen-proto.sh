#!/usr/bin/env bash

search_dir=pkg/grpc/proto

for entry in "$search_dir"/*
do
  protoFile=`ls $entry/*.proto`
  for pFile in $protoFile;
  do
    echo "Generating protogen files for $pFile"
    protoc -I. -Ivendor/ --go_out=paths=source_relative,plugins=grpc:. $pFile
  done
done