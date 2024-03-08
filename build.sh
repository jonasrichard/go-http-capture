#!/bin/bash

docker build . --tag capture:build --platform linux/amd64

id=$(docker create capture:build)

docker cp $id:/src/target/capture capture-linux

docker rm -v $id
