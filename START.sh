#!/bin/bash
# START.sh -- all in one start script

# TODO: Kill and start Icecast
# TODO: Kill and start Liquidsoap

# Run configuration setup script
# This syntax '. ./SETUP.sh' is necessary so exports are set
# in the correct shell.
. ./scripts/config.sh

# Run the server application code
go run *.go