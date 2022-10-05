#!/bin/bash

# used during development

wd=$(pwd)

while true
do
    set +x
    echo ==========
    set -x
    cd $wd
    inotifywait -r -e modify .
    sleep 1
    [ -n "$pid" ] && kill $pid
    # go test -v || continue 
    go build || continue

    go run . example/pandemic.cmap example/pandemic.dot
    xdot example/pandemic.dot &

    pid=$!
done
