#!/bin/bash

sudo nohup python -OO server/server.py 8080 ./public/ > /dev/null &
