#!/bin/bash

echo Creating mock GOPATH
export GOPATH=~/go

mkdir -p $GOPATH/src/github.com/kittycash
ln -s $PWD $GOPATH/src/github.com/kittycash/teller
cd $GOPATH/src/github.com/kittycash/teller

echo Installing "'dep'"
mkdir -p ~/go/bin
export PATH=~/go/bin:$PATH
curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

echo Running "'dep ensure -v'"
dep ensure -v

for tag in $(./docker-tags.sh) ; do
  docker build --pull --cache-from kittycash/teller --tag "$tag" . || exit 1
done
