#!/bin/bash

nohup python3 -OO server/server.py 8080 ./public/ > /dev/null &
