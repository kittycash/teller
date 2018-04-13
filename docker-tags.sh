#!/bin/bash

if [ "$TRAVIS_BRANCH" == "master" ]; then
  echo kittycash/teller:latest
fi

if [ "$TRAVIS_BRANCH" == "develop" ]; then
  echo kittycash/teller:latest-develop
fi

echo kittycash/teller:$(git rev-parse --short HEAD)

git tag --points-at HEAD | \
  sed -e 's/\(.*\)/kittycash\/teller:\1/g'
