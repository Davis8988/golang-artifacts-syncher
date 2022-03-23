
# Builds the golang artifacts syncher docker image


export GOPROXY=http://artifactory.esl.corp.elbit.co.il/artifactory/GO
export SYNCHER_BUILD_VERSION=${SYNCHER_BUILD_VERSION:-1.0.0}


DOCKER_IMAGE_TAG=artifactory.esl.corp.elbit.co.il/aerospace-simulators-devops-docker/golang/artifacts-syncher/go-1.18-alpine:${SYNCHER_BUILD_VERSION}
cmnd="docker build .. -t $DOCKER_IMAGE_TAG --add-host=artifactory.esl.corp.elbit.co.il:10.0.50.35 --build-arg SYNCHER_BUILD_VERSION=${SYNCHER_BUILD_VERSION}"
echo Executing: $cmnd
eval $cmnd
if [ "$?" != "0" ]; then echo "" && echo Error - Failed during execution of: && echo $cmnd; fi
echo ""
echo "Success"
echo ""
