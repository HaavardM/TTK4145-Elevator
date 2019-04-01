#!/bin/bash

PROJECT_PATH=$GOPATH/src/github.com/TTK4145-students-2019

set +e
mkdir -p $PROJECT_PATH

cd $PROJECT_PATH

git clone https://github.com/TTK4145-students-2019/project-thefuturezebras.git

set -e


cd $PROJECT_PATH/project-thefuturezebras
go get

echo "Project ready!"


