#!/bin/bash

# Exit immediately upon error
set -eo pipefail

if [ $# -gt 0 ]
then
      echo "$(basename $0): No parameters allowed, $# given."
      exit 1
fi

cat <<END
"***************************************************************"
NOTE: If you need help determining configuration values to use,
installation documentation is available on GitHub:
https://github.com/kenellorando/cadence/wiki/Installation
***************************************************************

[1/5] Path to Music Directory
Set a path to a directory containing audio files (e.g. mp3, flac) to be played
on the radio. The target will be recursively searched.
END
read -p "      Music path: " CADENCE_PATH
while [ ! -d "$CADENCE_PATH" ]
do
      echo "Music path must point to a directory that exists and is readable."
      read -p "      Music path: " CADENCE_PATH
done
# We do need to use absolute paths here - Make sure they end up that way.
# realpath -s is used here instead of readlink -f to retain symlinks - Else,
# we'd automatically go to the destination and use that, even if the user
# changed the symlink and restarted Cadence!
# ... Despite this, symlinks inside CADENCE_PATH probably won't work, since
# they don't get mounted inside our containers. Not a lot we can do about that.
CADENCE_PATH=$(realpath -s "$CADENCE_PATH")

echo

cat <<END
[2/5] Stream Host Address
Set the stream host address. This may be a DNS name, public IP, or private IP.
Use localhost:8000 if your Cadence instance is meant for local use only.
Default: localhost:8000
END
read -p "      Stream address: " CADENCE_STREAM_HOST
if [ -z "$CADENCE_STREAM_HOST" ]
then
      echo "Streaming to localhost:8000."
      CADENCE_STREAM_HOST='localhost:8000'
fi


echo

cat <<END
[3/5] Rate Limiter Timeout
Set a rate limit timeout in integer seconds. This prevents the same listener
from requesting songs within the configured timeframe. Set to 0 to disable.
END
read -p "      Rate limit (0): " CADENCE_RATE
while ! [[ "$CADENCE_RATE" =~ ^\d*$ ]]
do
      echo "Rate limit must be an integer!"
      read -p "      Rate limit (0): " CADENCE_RATE
done
[ -z "$CADENCE_RATE" ] && CADENCE_RATE=0


echo

cat <<END
[4/5] Radio Service Password
Set a secure, unique service password. Input is hidden.
END
read -s -p "      Password: " CADENCE_PASS
while [ -z "$CADENCE_PASS" ]
do
      echo
      echo "Password cannot be empty!"
      read -s -p "      Password: " CADENCE_PASS
done

echo
echo

cat <<END
[5/5] Enable Reverse Proxy?
Do you want to enable a reverse proxy? Skip if you are broadcasting locally only
or have your own reverse proxy configured. Skip if you do not know what this means.
END
ENABLE_REVERSE_PROXY="UNSET"
while ! [[ "$ENABLE_REVERSE_PROXY" =~ ^[yYnN]$ ]] && [ -n "$ENABLE_REVERSE_PROXY" ]
do
      read -n1 -p "      [y/N]: " ENABLE_REVERSE_PROXY
      echo
done

if [[ "$ENABLE_REVERSE_PROXY" =~ ^([yY])$ ]]
then
      echo "Please provide the domain name you will use for Cadence UI."
      read -p "      Web UI Domain: " CADENCE_WEB_HOST
      while [ -z "$CADENCE_WEB_HOST" ]
      do
            echo "Web UI Domain cannot be empty!"
            read -p "      Web UI Domain: " CADENCE_WEB_HOST
      done
else
      echo "No reverse proxy will be configured."
fi

SCRIPT_DIR="$(dirname $(readlink -f $0))"
cd $SCRIPT_DIR

cp ./config/cadence.env.example ./config/cadence.env
cp ./config/icecast.xml.example ./config/icecast.xml
cp ./config/liquidsoap.liq.example ./config/liquidsoap.liq
cp ./config/nginx.conf.example ./config/nginx.conf
cp ./docker-compose.yml.example ./docker-compose.yml

if [[ "$ENABLE_REVERSE_PROXY" =~ ^([yY])$ ]]
then
      awk -i inplace -v "c=$(cat ./nginx-compose-section.yml)" \
          '{gsub(/NGINX_CONFIG_SECTION/,c)}1' docker-compose.yml
else
      sed -i 's|NGINX_CONFIG_SECTION||g' ./docker-compose.yml
fi

sed -i 's|CADENCE_PASS_EXAMPLE|'"$CADENCE_PASS"'|g' ./config/cadence.env
sed -i 's|CADENCE_PASS_EXAMPLE|'"$CADENCE_PASS"'|g' ./config/icecast.xml
sed -i 's|CADENCE_PASS_EXAMPLE|'"$CADENCE_PASS"'|g' ./config/liquidsoap.liq
sed -i 's|CADENCE_RATE_EXAMPLE|'"$CADENCE_RATE"'|g' ./config/cadence.env
sed -i 's|CADENCE_STREAM_HOST_EXAMPLE|'"$CADENCE_STREAM_HOST"'|g' ./config/icecast.xml
sed -i 's|CADENCE_PATH_EXAMPLE|'"$CADENCE_PATH"'|g' ./config/cadence.env
sed -i 's|CADENCE_PATH_EXAMPLE|'"$CADENCE_PATH"'|g' ./config/liquidsoap.liq
sed -i 's|CADENCE_STREAM_HOST_EXAMPLE|'"$CADENCE_STREAM_HOST"'|g' ./config/nginx.conf
sed -i 's|CADENCE_WEB_HOST_EXAMPLE|'"$CADENCE_WEB_HOST"'|g' ./config/nginx.conf
sed -i 's|CADENCE_PATH_EXAMPLE|'"$CADENCE_PATH"'|g' ./docker-compose.yml

echo ""
echo "Configuration completed."

docker compose down
docker compose pull
docker compose up
