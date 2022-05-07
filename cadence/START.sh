#!/bin/bash
# START.sh -- all in one start script

# Run configuration setup script
# This syntax '. ./SETUP.sh' is necessary so exports are set
# in the correct shell.
. ./scripts/config.sh

# Run the server application code
go run server/*.go
