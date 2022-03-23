
FROM artifactory.esl.corp.elbit.co.il/aerospace-simulators-devops-docker/golang:go-1.18-alpine-curl-git AS builder
ARG EXECUTABLE_APP_PATH=/app/bin/golang-artifacts-syncher
ARG SYNCHER_BUILD_VERSION=$SYNCHER_BUILD_VERSION

# RUN apk --no-cache add networkmanager git curl

ENV GOPROXY=http://artifactory.esl.corp.elbit.co.il/artifactory/GO
# ENV GONOSUMDB=github.com/*,golang.org/*,sum.golang.org/*,go.starlark.net/*,go.starlark.net,gopkg.in/*,rsc.io/*,go.etcd.io/*,go.uber.org/*,cloud.google.com/*,google.golang.org/*,go.opencensus.io,honnef.co/*,dmitri.shuralyov.com/*,9fans.net/*,mvdan.cc/*
ENV GONOSUMDB=github.com/*
ENV GONOSUMDB=$GONOSUMDB,golang.org/*
ENV GONOSUMDB=$GONOSUMDB,sum.golang.org/*
ENV GONOSUMDB=$GONOSUMDB,go.starlark.net/*
ENV GONOSUMDB=$GONOSUMDB,go.starlark.net
ENV GONOSUMDB=$GONOSUMDB,gopkg.in/*
ENV GONOSUMDB=$GONOSUMDB,rsc.io/*
ENV GONOSUMDB=$GONOSUMDB,go.etcd.io/*
ENV GONOSUMDB=$GONOSUMDB,go.uber.org/*
ENV GONOSUMDB=$GONOSUMDB,cloud.google.com/*
ENV GONOSUMDB=$GONOSUMDB,google.golang.org/*
ENV GONOSUMDB=$GONOSUMDB,go.opencensus.io
ENV GONOSUMDB=$GONOSUMDB,honnef.co/*
ENV GONOSUMDB=$GONOSUMDB,dmitri.shuralyov.com/*
ENV GONOSUMDB=$GONOSUMDB,9fans.net/*
ENV GONOSUMDB=$GONOSUMDB,mvdan.cc/*


WORKDIR /app

COPY . .

RUN chmod +x ./docker/entrypoint.sh && \ 
    echo Executing: go build -buildmode=exe -o $EXECUTABLE_APP_PATH -ldflags="-s -w -X 'main.BuildVersion=$SYNCHER_BUILD_VERSION'" src/golang-artifacts-syncher.go && \
    go build -buildmode=exe -o $EXECUTABLE_APP_PATH -ldflags="-s -w -X 'main.BuildVersion=$SYNCHER_BUILD_VERSION'" src/golang-artifacts-syncher.go

# Runtime image:
FROM artifactory.esl.corp.elbit.co.il/aerospace-simulators-devops-docker/alpine:3.14.4-curl-git

WORKDIR /app
COPY --from=builder "/app/bin/golang-artifacts-syncher" "/usr/bin/"
COPY ./docker/entrypoint.sh "/app/bin/"

CMD ["sh", "/app/bin/entrypoint.sh"]

