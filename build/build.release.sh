#!/bin/bash

echo Setting USED_DOCKER_REPO="aerospace-simulators-devops-docker"
export USED_DOCKER_REPO="aerospace-simulators-devops-docker"

echo Adding permissions to: ./build.sh
chmod +x ./build.sh

echo Executing: ./build.sh
./build.sh
