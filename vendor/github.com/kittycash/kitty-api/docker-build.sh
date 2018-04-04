#!/bin/bash

for tag in $(./docker-tags.sh) ; do
  docker build --pull --cache-from kittycash/kitty-api --tag "$tag" . || exit 1
done
