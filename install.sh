#!/bin/bash

echo "***************************************************************"
echo "NOTE: If you need help determining configuration values to use,"
echo "installation documentation is available on GitHub:"
echo "https://github.com/kenellorando/cadence/wiki/Installation"
echo "***************************************************************"
echo ""
echo "[1/5] Absolute Path to Music"
echo "Set an absolute path to a directory containing audio files (e.g. mp3, flac)"
echo "to be played on the radio. The target is not recursively searched."
read -p "      Music path: " CADENCE_PATH
echo ""
echo "[2/5] Stream Host Address"
echo "Set the stream host address. This may be a DNS name, public IP, or private IP."
echo "Use localhost:8000 if your Cadence instance is meant for local use only."
read -p "      Stream address: " CADENCE_STREAM_HOST
echo ""
echo "[3/5] Rate Limiter Timeout"
echo "Set a rate limit timeout in integer seconds. This prevents the same listener"
echo "from requesting songs within the configured timeframe. Set to 0 to disable."
read -p "      Rate limit: " CADENCE_RATE
echo ""
echo "[4/5] Radio Service Password"
echo "Set a secure, unique service password. Input is hidden."
read -s -p "      Password: " CADENCE_PASS
echo ""
echo ""
echo "[5/5] Enable Reverse Proxy?"
echo "Do you want to enable a reverse proxy? Skip if you are broadcasting locally only"
echo "or have your own reverse proxy configured. Skip if you do not know what this means."
read -p "      [y/N]: " ENABLE_REVERSE_PROXY

if [[ "$ENABLE_REVERSE_PROXY" =~ ^([yY])$ ]]
then
      echo "Please provide the domain name you will use for Cadence UI."
      read -p "      Web UI Domain: " CADENCE_WEB_HOST
else
      echo "No reverse proxy will be configured."
fi

cp ./config/cadence.env.example ./config/cadence.env
cp ./config/icecast.xml.example ./config/icecast.xml
cp ./config/liquidsoap.liq.example ./config/liquidsoap.liq
cp ./config/nginx.conf.example ./config/nginx.conf
cp ./docker-compose.yml.example ./docker-compose.yml

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

if [[ "$ENABLE_REVERSE_PROXY" =~ ^([yY])$ ]]
then
      docker compose --profile nginx up 
else
      docker compose up
fi
