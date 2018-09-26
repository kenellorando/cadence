#!/bin/bash

sudo nohup python -OO server/server.py 80 ./public/ > /dev/null &
