#!/bin/bash

if [ "$TRAVIS_BRANCH" == "master" ]; then
  echo kittycash/kitty-api:latest
fi

if [ "$TRAVIS_BRANCH" == "develop" ]; then
  echo kittycash/kitty-api:latest-develop
fi

echo kittycash/kitty-api:$(git rev-parse --short HEAD)

git tag --points-at HEAD | \
  sed -e 's/\(.*\)/kittycash\/kitty-api:\1/g'
