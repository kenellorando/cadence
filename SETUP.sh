#!/bin/bash

# Prompt for environment variables that have defaults set by the server first
# Webserver
read -p "Log level (default: 5): " CSERVER_LOGLEVEL
CSERVER_LOGLEVEL=${CSERVER_LOGLEVEL:-5}
read -p "Server domain (default: localhost): " CSERVER_DOMAIN
CSERVER_DOMAIN=${CSERVER_DOMAIN:-localhost}
read -p "Server port (default: :8080): " CSERVER_PORT
CSERVER_PORT=${CSERVER_PORT:-:8080}
read -p "Music directory absolute path (default: ~/cadence-music/): " CSERVER_MUSIC_DIR
CSERVER_MUSIC_DIR=${CSERVER_MUSIC_DIR:-~/cadence_music}
# Database
read -p "Database domain (default: localhost): " CSERVER_DB_DOMAIN
CSERVER_DB_DOMAIN=${CSERVER_DB_DOMAIN:-localhost}
read -p "Database port (default: 5432): " CSERVER_DB_PORT
CSERVER_DB_PORT=${CSERVER_DB_PORT:-5432}
read -p "Database name (default: cadence): " CSERVER_DB_NAME
CSERVER_DB_NAME=${CSERVER_DB_NAME:-cadence}
read -p "Database SSL mode (default: disable): " CSERVER_DB_SSL
CSERVER_DB_SSL=${CSERVER_DB_SSL:-disable}
read -p "Database driver (default: postgres): " CSERVER_DB_DRIVER
CSERVER_DB_DRIVER=${CSERVER_DB_DRIVER:-postgres}
read -p "Database user (default: postgres): " CSERVER_DB_USER
CSERVER_DB_USER=${CSERVER_DB_USER:-postgres}
# Table schema
read -p "Database table (default: aria): " CSERVER_DB_TABLE
CSERVER_DB_TABLE=${CSERVER_DB_TABLE:-aria}


# Prompt for essential user-created environment variables next
echo
read -s -p "Database password: " CSERVER_DB_PASS

echo
set | grep 'CSERVER_'