#!/bin/bash

# Exit immediately upon error, and treat an unset variable as one.
set -euo pipefail

CADENCE_WEB_HOST=""
ENABLE_REVERSE_PROXY=""

if [ $# -gt 0 ]
then
      echo "$(basename $0): No parameters allowed, $# given."
      exit 1
fi

# Everything below ends in "docker compose up", so check that will work before
# asking for any configuration. Failing after four prompts wastes the answers.
if ! command -v docker > /dev/null 2>&1
then
      echo "Docker is not installed, or is not on PATH."
      echo "Install Docker Engine: https://docs.docker.com/engine/install/"
      exit 1
fi
if ! docker compose version > /dev/null 2>&1
then
      echo "Docker Compose V2 is not available ('docker compose version' failed)."
      echo "Install Compose V2: https://docs.docker.com/compose/install/"
      exit 1
fi
if ! docker info > /dev/null 2>&1
then
      echo "The Docker daemon is not responding. Start Docker and try again."
      exit 1
fi

cat <<END
"***************************************************************"
NOTE: If you need help determining configuration values to use,
installation documentation is available on GitHub:
https://github.com/kenellorando/cadence/wiki/Installation
***************************************************************

[1/4] Path to Music Directory
Set a path to a directory containing audio files (e.g. mp3, flac) to be played
on the radio. The target will be recursively searched.
END
read -ep "      Music path: " CADENCE_PATH
while [ ! -d "$CADENCE_PATH" ]
do
      echo "Music path must point to a directory that exists and is readable."
      read -ep "      Music path: " CADENCE_PATH
done
# We do need to use absolute paths here - Make sure they end up that way.
# realpath -s is used here instead of readlink -f to retain symlinks - Else,
# we'd automatically go to the destination and use that, even if the user
# changed the symlink and restarted Cadence!
# ... Despite this, symlinks inside CADENCE_PATH probably won't work, since
# they don't get mounted inside our containers. Not a lot we can do about that.
CADENCE_PATH=$(realpath -s "$CADENCE_PATH")

# The station comes up either way, but silently empty is a confusing first
# impression, so say so now rather than leaving it to be discovered.
if [ -z "$(find "$CADENCE_PATH" -type f \( -iname '*.mp3' -o -iname '*.flac' -o -iname '*.ogg' \) -print -quit 2>/dev/null)" ]
then
      echo
      echo "      Warning: no .mp3, .flac or .ogg files were found under this path."
      echo "      Cadence will start, but the library and radio will be empty."
fi

echo

cat <<END
[2/4] Rate Limiter Timeout
Set a rate limit timeout in integer seconds. This prevents the same listener
from requesting songs within the configured timeframe. Set to 0 to disable.
END
read -p "      Rate limit (0): " CADENCE_RATE
while ! [[ "$CADENCE_RATE" =~ ^[0-9]*$ ]]
do
      echo "Rate limit must be an integer!"
      read -p "      Rate limit (0): " CADENCE_RATE
done
[ -z "$CADENCE_RATE" ] && CADENCE_RATE=0


echo

cat <<END
[3/4] Radio Service Password
Set a secure, unique service password. Input is hidden.
END
CADENCE_PASS=""
CADENCE_PASS_CONFIRM=""
while true
do
      read -s -p "      Password: " CADENCE_PASS
      echo
      if [ -z "$CADENCE_PASS" ]
      then
            echo "Password cannot be empty!"
            continue
      fi
      read -s -p "      Confirm password: " CADENCE_PASS_CONFIRM
      echo
      if [ "$CADENCE_PASS" = "$CADENCE_PASS_CONFIRM" ]
      then
            break
      fi
      echo "Passwords did not match. Try again."
done

echo
echo

cat <<END
[4/4] Enable Reverse Proxy?
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

# sed gives \, & and our | delimiter special meaning on the replacement side.
# An unescaped & expands to the whole match, so a password like "p@ss&word"
# would silently land in the config as "p@ssCADENCE_PASS_EXAMPLEword", and a
# password containing | would abort the script mid-configuration.
sed_escape() {
      printf '%s' "$1" | sed 's/[\\&|]/\\&/g'
}

# GNU sed -i takes no argument while BSD/macOS sed requires one, so the same
# invocation cannot work on both. Edit through a temporary file instead.
replace_in_file() {
      local placeholder="$1" value="$2" file="$3"
      sed "s|${placeholder}|$(sed_escape "$value")|g" "$file" > "${file}.tmp"
      mv "${file}.tmp" "$file"
}

cp ./config/cadence.env.example ./config/cadence.env
cp ./config/liquidsoap.liq.example ./config/liquidsoap.liq
cp ./config/nginx.conf.example ./config/nginx.conf

if [[ "$ENABLE_REVERSE_PROXY" =~ ^([yY])$ ]]
then
      awk -v "c=$(cat ./nginx-compose-section.yml)" \
          '{gsub(/NGINX_CONFIG_SECTION/,c)}1' ./docker-compose.yml.example > ./docker-compose.yml
else
      sed -e 's|NGINX_CONFIG_SECTION||g' ./docker-compose.yml.example > ./docker-compose.yml
fi

replace_in_file CADENCE_PASS_EXAMPLE "$CADENCE_PASS" ./config/cadence.env
replace_in_file CADENCE_PASS_EXAMPLE "$CADENCE_PASS" ./config/liquidsoap.liq
replace_in_file CADENCE_RATE_EXAMPLE "$CADENCE_RATE" ./config/cadence.env
replace_in_file CADENCE_PATH_EXAMPLE "$CADENCE_PATH" ./config/cadence.env
replace_in_file CADENCE_PATH_EXAMPLE "$CADENCE_PATH" ./config/liquidsoap.liq
replace_in_file CADENCE_WEB_HOST_EXAMPLE "$CADENCE_WEB_HOST" ./config/nginx.conf
replace_in_file CADENCE_PATH_EXAMPLE "$CADENCE_PATH" ./docker-compose.yml

echo ""
echo "Configuration completed. Starting your station..."
echo ""

docker compose down --remove-orphans
docker compose pull
docker compose up -d

cat <<END

Cadence is running.

      Web UI:  http://localhost:8080

The music library is read in the background. A large library takes a while, and
the search tab will say so until it has finished.

      Follow the logs:  docker compose logs -f
      Stop the station: docker compose down
      Reconfigure:      ./install.sh
END
