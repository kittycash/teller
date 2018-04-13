#!/bin/bash

if [ ! -f ~/.netrc ] ; then
  echo "Configuring private dependency access"
  # configuring private dependency access:
  #https://docs.travis-ci.com/user/languages/go/#Installing-Private-Dependencies
  echo "machine github.com
  login $GITHUB_USER
  password $GITHUB_KEY
" > ~/.netrc
  chmod 600 ~/.netrc
fi

echo Creating mock GOPATH
export GOPATH=~/go

if [ ! -e $GOPATH/src/github.com/kittycash/teller ] ; then

  mkdir -p $GOPATH/src/github.com/kittycash
  ln -s $PWD $GOPATH/src/github.com/kittycash/teller
  cd $GOPATH/src/github.com/kittycash/teller

fi

if [ ! $(which dep) ] ; then

  echo Installing "'dep'"
  mkdir -p ~/go/bin
  export PATH=~/go/bin:$PATH
  curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

fi

echo Running "'dep ensure -v'"
dep ensure -v || exit 1

echo "Running 'make test'"
make test || exit 1

for tag in $(./docker-tags.sh) ; do
  docker build --pull --cache-from kittycash/teller --tag "$tag" . || exit 1
done
