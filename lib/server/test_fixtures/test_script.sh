#!/bin/bash

if [ -z "$1" ]; then
    echo "Usage: $0 <argument>"
    exit 1
fi

for i in {1..5}; do
    echo "Line $i: $1"
    sleep 1
done
