#/bin/bash
docker kill $(docker ps -a -q)
