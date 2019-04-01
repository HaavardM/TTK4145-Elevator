#!/bin/sh

EXIT_CODE=1
while [ $EXIT_CODE -gt 0 ]; do 
    $@
    EXIT_CODE=$?
    sleep 1s
done
