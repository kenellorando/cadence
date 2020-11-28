#!/bin/bash
# SETUP.sh -- sets all environment variables used as 
# configuration in the Cadence server.
# Default config values can also be set here.

# Use dot syntax to ensure configs are exported to the right shell.
# Usage: . ./SETUP.sh

echo "SETUP.sh"

# Functions
#############################################################

# Use SetEnvVar for any value that does not need to be hidden.
# Optionally allows a default value to be set.
function SetEnvVar() {
    ENVVAR=$1
    DEFAULT=$2
    if [[ -z $DEFAULT ]]
    then 
        PROMPT="$ENVVAR: "
    else 
        PROMPT="$ENVVAR (default: $DEFAULT): "
    fi

    read -p "$PROMPT" INPUT

    if [[ -z $INPUT ]]
    then
        export $ENVVAR=$DEFAULT
    else
        export $ENVVAR=$INPUT
    fi
}

# Use SetSecretEnvVar to hide input.
# Does not allow setting default values.
function SetSecretEnvVar() {
    INPUT=null
    ENVVAR=$1
    PROMPT="$ENVVAR: "

    read -s -p "$ENVVAR: " INPUT

    if [[ -z $INPUT ]]
    then
        echo -e "\n$ENVVAR has no default and cannot be left blank."
        SetSecretEnvVar "$1"
    else
	export $ENVVAR=$INPUT    
    fi
}


# Set variables here
##############################################################

# List of environment variables that have defaults.
# These need to be set with SetEnvVar()
SetEnvVar "CSERVER_ROOTPATH" "/home/cadence/cadence/"
SetEnvVar "CSERVER_LOGLEVEL" "5"
SetEnvVar "CSERVER_DOMAIN" "localhost"
SetEnvVar "CSERVER_PORT" ":8080"
SetEnvVar "CSERVER_MUSIC_DIR" "/home/cadence/cadence_music/"
SetEnvVar "CSERVER_SOURCEADDRESS" "localhost:1234"
SetEnvVar "CSERVER_DB_HOST" "localhost"
SetEnvVar "CSERVER_DB_PORT" "5432"
SetEnvVar "CSERVER_DB_NAME" "cadence"
SetEnvVar "CSERVER_DB_SSLMODE" "disable"
SetEnvVar "CSERVER_DB_DRIVER" "postgres"
SetEnvVar "CSERVER_DB_USER" "postgres"
SetEnvVar "CSERVER_DB_TABLE" "aria"
SetEnvVar "CSERVER_WHITELIST_PATH" "/home/cadence/cadence/whitelist.txt"
# List of environment variables with no defaults.
# Does not necessarily need to be used with SetSecretEnvVar(),
# though most envvars here would make sense to hide the input.
SetSecretEnvVar "CSERVER_DB_PASS"

##############################################################

touch $CSERVER_WHITELIST_PATH
chmod 755 $CSERVER_WHITELIST_PATH

echo -e "\nSETUP.sh completed."
