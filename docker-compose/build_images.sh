#!/bin/bash

GATEWAY_DOCKERFILE="../gateway/dockerfile"
FILES_DOCKERFILE="../files/dockerfile"

NEW_TAG="latest"
OLD_TAG="latest"

docker rmi $(docker images -q fileshare/gateway:$OLD_TAG) 2>/dev/null
docker rmi $(docker images -q fileshare/files:$OLD_TAG) 2>/dev/null

docker build -t fileshare/gateway:$NEW_TAG -f $GATEWAY_DOCKERFILE ../gateway
docker build -t fileshare/files:$NEW_TAG -f $FILES_DOCKERFILE ../files
