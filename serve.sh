#!/bin/bash

# Kill existing instances of the server
echo "[1] Killing existing instances of Cadence..."
pkill python3
echo

echo "[2] Starting server..."
# Start server
nohup python3 -OO server/server.py 8080 ./public/ > /dev/null &

echo
echo "Webserver started!"
