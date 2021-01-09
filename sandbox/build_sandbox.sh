#!/bin/bash

gcc src/sandbox.c -static -static-libgcc -static-libstdc++ -o sandbox
sudo docker build . -t sandbox:latest

rm sandbox