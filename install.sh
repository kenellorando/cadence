#!/bin/bash

read -p "Music directory absolute path (/music/): " C_PATH
read -p "Stream hostname (localhost): " C_HOST
read -p "Rate limit timeout (180):  " C_RATE
read -s -p "Service password: " C_PASS

read -p "Do you have DNS? [y/n]: " HAS_DNS

cp ./config/cadence.env.example ./config/cadence.env
cp ./config/icecast.xml.example ./config/icecast.xml
cp ./config/liquidsoap.liq.example ./config/liquidsoap.liq
cp ./config/nginx.conf.example ./config/nginx.conf

