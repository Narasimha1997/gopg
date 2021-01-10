#!/bin/bash

pushd ../client
    mkdir -p ../bin
    go build -o ../bin/gopg-client
    echo "Built bin/gopg-client"
popd