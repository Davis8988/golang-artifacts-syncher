#!/bin/bash

# Build golang syncher project
# 
# Version number is taken from the most recent anotated tag that should have the format: '_._._'
#   using command: 'git describe --abbrev=0'
# 
# Env: 'USED_DOCKER_REPO' must be defined or passed as 1st param
#  it must be one of the following valid values: "aerospace-simulators-devops-docker-integ", "aerospace-simulators-devops-docker"
#
#
# By David Yair [E030331]
# 

USED_DOCKER_REPO=${1:-$USED_DOCKER_REPO}  # Either use 1st arg or existing env: 'USED_DOCKER_REPO'
if [ "$USED_DOCKER_REPO" != "aerospace-simulators-devops-docker-integ" ] && [ "$USED_DOCKER_REPO" != "aerospace-simulators-devops-docker" ]; then echo '' && echo "Error - env USED_DOCKER_REPO has value of: '$USED_DOCKER_REPO'. Please specify a docker repo from these valid values: 'aerospace-simulators-devops-docker-integ', 'aerospace-simulators-devops-docker'"; exit 1; fi

export GOPROXY=${GOPROXY:-http://artifactory.esl.corp.elbit.co.il/artifactory/GO}
PROJ_NAME="Golang-Artifacts-Syncher"
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
REPO_ROOT_DIR=$(git rev-parse --show-toplevel)
_getLatestTagCmnd="git describe --abbrev=0"
_sleep_wait_for_container_healthy_sec=4
_max_retry_count_wait_for_container_healthy=10


echo '' && echo '------- Preparing Envs -------' && echo ''
echo Executing: ${_getLatestTagCmnd}
export SYNCHER_BUILD_VERSION=$(${_getLatestTagCmnd})
if [ "$?" != "0" ]; then echo '' && echo "Error - Failed to get latest tag using command: '${_getLatestTagCmnd}'"; exit 1; fi
if [ -z "$SYNCHER_BUILD_VERSION" ]; then echo '' && echo "Error - env 'SYNCHER_BUILD_VERSION' is empty. Failed to get latest tag using command: '${_getLatestTagCmnd}'"; exit 1; fi
if [ -z "$USED_DOCKER_REPO" ]; then echo '' && echo "Error - env 'USED_DOCKER_REPO' is empty. Please specify a docker repo from these valid values: 'aerospace-simulators-devops-docker-integ', 'aerospace-simulators-devops-docker'"; exit 1; fi
echo "Syncher build version: ${SYNCHER_BUILD_VERSION}"

DOCKER_REGISTERY_HOSTNAME="artifactory.esl.corp.elbit.co.il"
SYNCHER_DOCKER_IMAGE_NAME="${DOCKER_REGISTERY_HOSTNAME}/${USED_DOCKER_REPO}/golang/artifacts-syncher/go-1.18-alpine"
SYNCHER_DOCKER_IMAGE_FULL_NAME="${SYNCHER_DOCKER_IMAGE_NAME}:${SYNCHER_BUILD_VERSION}"
CLEAN_DOCKER_IMAGES_REPO_URL = "http://${DOCKER_REGISTERY_HOSTNAME}/artifactory/api/docker/${USED_DOCKER_REPO}/v2/golang/artifacts-syncher/go-1.18-alpine/tags/list"


# Checks:
if [ -z "$SCRIPT_DIR" ]; then echo "Error - Missing env var: SCRIPT_DIR that should point to this script's dir"; exit 1; fi
if [ -z "$REPO_ROOT_DIR" ]; then echo "Error - Missing env var: REPO_ROOT_DIR that should point to root level of ${PROJ_NAME} repository dir"; exit 1; fi
if [ ! -d "${REPO_ROOT_DIR}/src" ]; then echo "Error - Missing dir: ${REPO_ROOT_DIR}/src"; exit 1; fi



# Docker build
echo '' && echo '------- Docker Image Build -------' && echo ''
_dockerBuildCmnd="docker build .. -t ${SYNCHER_DOCKER_IMAGE_FULL_NAME} --pull --add-host=${DOCKER_REGISTERY_HOSTNAME}:10.0.50.35 --build-arg SYNCHER_BUILD_VERSION=$SYNCHER_BUILD_VERSION"
echo Executing: ${_dockerBuildCmnd}
eval ${_dockerBuildCmnd}
if [ "$?" != "0" ]; then echo '' && echo "Error - Failure during execution of docker build command: '${_dockerBuildCmnd}'"; exit 1; fi
echo 'OK'
echo 'Success - Finished building docker image: ${SYNCHER_DOCKER_IMAGE_FULL_NAME}'
echo ''

# Docker Run
_containerName=$(echo "${SYNCHER_BUILD_VERSION}_${SYNCHER_DOCKER_IMAGE_FULL_NAME}" | sed "s/:/_/g")  # 'g' -> Replace all ':' with: '_'  
_containerName=$(echo "${_containerName}" | sed "s/\//_/g")  # 'g' -> Replace all '/' with: '_'  
echo "Running new container from image: '${SYNCHER_DOCKER_IMAGE_FULL_NAME}'"
echo " with name: ${_containerName}"

for i in $(docker ps -aq -fname=${_containerName}); do echo Stopping and removing container: ${_containerName} && docker rm -f ${i}; done

_dockerRunCmnd="docker run -d --name=${_containerName} ${SYNCHER_DOCKER_IMAGE_FULL_NAME}"
echo Executing: ${_dockerRunCmnd}
eval ${_dockerRunCmnd}
if [ "$?" != "0" ]; then echo '' && echo "Error - Failure during execution of docker run command: '${_dockerRunCmnd}'"; exit 1; fi
echo 'OK'
echo ''

# Docker Health
echo inspecting container: ${_containerName}; echo; docker ps
echo -e "\n ## Waiting for container ${_containerName} to be healthy ##\n"
echo "Sleeping for 1 seconds"
sleep 1

echo Executing: docker inspect ${_containerName} --format="{{ .State.Health.Status }}"
_container_health_status=$(docker inspect ${_containerName} --format="{{ .State.Health.Status }}")
echo "Container health: ${_container_health_status}"
while [ "${_container_health_status}" != "healthy" ]
do
    echo "Sleeping for ${_sleep_wait_for_container_healthy_sec} seconds"
    sleep ${_sleep_wait_for_container_healthy_sec}
    echo ''
    echo Executing: docker inspect ${_containerName} --format="{{ .State.Health.Status }}"
    _container_health_status=$(docker inspect ${_containerName} --format="{{ .State.Health.Status }}")
    echo "Container health: ${_container_health_status}"
done
echo ''
if [ "${_container_health_status}" != "healthy" ]; then echo '' && Error - Container is not healthy: exit 1; fi

# Docker Logs
echo "Reading container logs: "
_dockerGetContainerLogsCmnd="docker logs ${_containerName}"
echo Executing: ${_dockerGetContainerLogsCmnd}
eval ${_dockerGetContainerLogsCmnd}
echo ''
if [ "$?" != "0" ]; then echo "Error - Failure during execution of docker run command: '${_dockerGetContainerLogsCmnd}'"; exit 1; fi

# Docker stop & remove
for i in $(docker ps -aq -fname=${_containerName}); do echo Stopping and removing container: ${_containerName} && docker rm -f ${i}; done

echo OK
echo ''
echo Success - Finished building docker image: ${SYNCHER_DOCKER_IMAGE_FULL_NAME}
echo ''
echo Done
echo Finished
echo ''
