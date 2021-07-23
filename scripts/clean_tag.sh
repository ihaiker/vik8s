#!/bin/bash

tag=$1
echo "delete origin branch release-$tag"
git push origin --delete release-$tag
if [ $? -gt 0 ]; then
    echo "branch not found"
fi

echo "delete origin tag $tag"
git push --delete origin $tag
if [ $? -gt 0 ]; then
    echo "origin tag not found"
fi

echo "delete local tag $tag"
git tag -d $tag
if [ $? -gt 0 ]; then
    echo "origin tag not found"
fi
