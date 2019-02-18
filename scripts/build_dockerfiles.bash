#!/bin/bash

TAG=$(git rev-parse HEAD)

docker build -t thefuturezebras/project:$TAG ..
docker push thefuturezebras/project:$TAG
