#!/bin/bash

# Builds the golang artifacts syncher docker image
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
REPO_ROOT_DIR=$(git rev-parse --show-toplevel)

export GOPROXY=${GOPROXY:-http://artifactory.esl.corp.elbit.co.il/artifactory/GO}

if [ -z "${SYNCHER_BUILD_VERSION}" ]; then echo Error - Missing env: SYNCHER_BUILD_VERSION && echo Cannot build syncher docker image without it && exit 1; fi

export SYNCHER_DOCKER_IMAGE_FULL_NAME=${SYNCHER_DOCKER_IMAGE_FULL_NAME:-artifactory.esl.corp.elbit.co.il/aerospace-simulators-devops-docker/golang/artifacts-syncher/go-1.18-alpine:${SYNCHER_BUILD_VERSION}}

echo SYNCHER_DOCKER_IMAGE_FULL_NAME=${SYNCHER_DOCKER_IMAGE_FULL_NAME}

cmnd="docker build \"${REPO_ROOT_DIR}\" -t $SYNCHER_DOCKER_IMAGE_FULL_NAME --add-host=artifactory.esl.corp.elbit.co.il:10.0.50.35 --build-arg SYNCHER_BUILD_VERSION=${SYNCHER_BUILD_VERSION}"
echo Executing: $cmnd
eval $cmnd
if [ "$?" != "0" ]; then echo "" && echo Error - Failed during execution of: && echo $cmnd && exit 1; fi
echo ""
echo "Success"
echo ""
