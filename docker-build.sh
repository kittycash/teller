#!/bin/bash

for tag in $(./docker-tags.sh) ; do
  docker build --pull --cache-from kittycash/teller --tag "$tag" . || exit 1
done
