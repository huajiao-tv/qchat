#!/bin/bash

paramCount=$#
buildPre=false
build=false
echo "usage: ./test.sh (start service)|./test.sh 1 (rebuild keeper,use to change conf)|./test.sh 2(rebuild qchat,use to change code)|./test.sh 1 2 (rebuild all)"
if [ $paramCount = 0 ] || [ $paramCount -gt 2 ];
then
  echo "no rebuild param or too many params,just up"
elif [ $paramCount = 1 ]
then
  if [ 1 = $1 ]
  then
    buildPro=true
  else
    build=true
  fi
else
  buildPre=true
  build=true
fi
if $buildPre
then
  echo "rebuild pre"
  docker-compose -f docker-compose-pre.yml down && docker-compose -f docker-compose-pre.yml build &&  docker-compose -f docker-compose-pre.yml up -d
else
  docker-compose -f docker-compose-pre.yml up -d
fi

if $build
then
  echo "rebuild main"
  docker-compose build
  ./wait-for-it.sh 127.0.0.1:17000 -- docker-compose up
else
  ./wait-for-it.sh 127.0.0.1:17000 -- docker-compose up
fi
