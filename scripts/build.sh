#!/bin/bash

SANDBOX_COMPILER="cc"
GO_COMPILER=/usr/local/go
BUILD_SANDBOX=$1


function build_sandbox () {
    sandbox_path=$1

    pushd $sandbox_path
        $SANDBOX_COMPILER src/sandbox.c \
            -static -static-libgcc -static-libstdc++ \
            -o sandbox
        
        docker build . -t sandbox
        rm sandbox

    popd

    echo "Built sandbox docker, to provide better isolation, install runsc, run scripts/install_runsc.sh"

}

function build_backend () {
    #builds the docker image of the backend
    pushd src
        go build -o ../bin/gopg
    popd

    echo "Built gopg, "
}

pushd ../

    if [[ "$BUILD_SANDBOX" == "--docker" ]]; then
        docker export $(docker create busybox) --output="busybox.tar"
        docker build . -t gopg:latest

        #sandbox
        build_sandbox $PWD/sandbox
        exit
    fi

    #sandbox
    build_sandbox $PWD/sandbox

    mkdir -p bin/
    pushd src 
        go build -o ../bin/gopg
        echo "Built gopg server binary"
    popd

    pushd client
        go build -o ../bin/gopg-client
        echo "Built gopg client"
    popd

popd


