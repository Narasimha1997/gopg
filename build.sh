#!/bin/bash

mkdir -p bin/

echo "Building gopg"

pushd src
    go build -o ../bin/gopg
popd 

pushd client
    go build -o ../bin/gopg-client
popd