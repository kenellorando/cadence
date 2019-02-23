#!/bin/bash
# START.sh -- 

# TODO: Kill and start Icecast
# TODO: Kill and start Liquidsoap

# Run configuration setup script
# This syntax '. ./SETUP.sh' is necessary so exports are set
# in the correct shell.
. ./SETUP.sh

# Run the server application code
go run *.go