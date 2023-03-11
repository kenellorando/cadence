#!/bin/bash

echo "[1/5] Music Directory Target"
echo "Set the absolute path of a directory containing audio files "
echo "(e.g. mp3, flac) meant for radio play. This target directory "
echo "will not be recursively searched. Example: /music/"
read -p "      Music path: " CADENCE_PATH
echo "=============================================================="
echo "[2/5] Stream Host Address"
echo "Set a stream host address for Cadence Icecast. This may be a "
echo "DNS name, public IP, or private IP. Set this to localhost:8000"
echo "if your Cadence instance is meant for local use only."
read -p "      Stream address: " CADENCE_HOST
echo "=============================================================="
echo "[3/5] Rate Limiter Timeout"
echo "Set a rate limit timeout in integer seconds. This prevents the "
echo "same listener from requesting songs within the configured "
echo "timeframe. Set to 0 to disable. Example: 180"
read -p "      Rate limit: " CADENCE_RATE
echo "=============================================================="
echo "[4/5] Radio Service Password"
echo "Set a secure, unique service password. Input is hidden."
read -s -p "      Password: " CADENCE_PASS
echo ""
echo "=============================================================="
echo "[5/5] Domain Names - LEAVE BLANK TO SKIP"
echo "OPTIONAL: if you are an advanced administrator routing DNS to "
echo "your Cadence stack, provide your domain names here. You will "
echo "be prompted for two domains: one for your Cadence Icecast "
echo "stream, one for the Cadence web UI. Subdomains are acceptable."
read -p "      Cadence Audio Stream Domain: " CADENCE_STREAM_DNS
read -p "      Cadence Web UI Domain: " CADENCE_WEB_DNS

cp ./config/cadence.env.example ./config/cadence.env
cp ./config/icecast.xml.example ./config/icecast.xml
cp ./config/liquidsoap.liq.example ./config/liquidsoap.liq
cp ./config/nginx.conf.example ./config/nginx.conf
cp ./docker-compose.yml.example ./docker-compose.yml

sed -i 's|CADENCE_PASS_EXAMPLE|'"$CADENCE_PASS"'|g' ./config/cadence.env
sed -i 's|CADENCE_PASS_EXAMPLE|'"$CADENCE_PASS"'|g' ./config/icecast.xml
sed -i 's|CADENCE_PASS_EXAMPLE|'"$CADENCE_PASS"'|g' ./config/liquidsoap.liq
sed -i 's|CADENCE_RATE_EXAMPLE|'"$CADENCE_RATE"'|g' ./config/cadence.env
sed -i 's|CADENCE_HOST_EXAMPLE|'"$CADENCE_HOST"'|g' ./config/icecast.xml
sed -i 's|CADENCE_PATH_EXAMPLE|'"$CADENCE_PATH"'|g' ./config/cadence.env
sed -i 's|CADENCE_PATH_EXAMPLE|'"$CADENCE_PATH"'|g' ./config/liquidsoap.liq
sed -i 's|CADENCE_STREAM_DNS_EXAMPLE|'"$CADENCE_STREAM_DNS"'|g' ./config/nginx.conf
sed -i 's|CADENCE_WEB_DNS_EXAMPLE|'"$CADENCE_WEB_DNS"'|g' ./config/nginx.conf
sed -i 's|CADENCE_PATH_EXAMPLE|'"$CADENCE_PATH"'|g' ./docker-compose.yml

echo "========================================="
echo "Configuration completed."
echo "If it does not begin automatically, run 'docker compose up' to start Cadence."
docker compose down
docker compose up
