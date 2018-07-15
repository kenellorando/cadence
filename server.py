#!/usr/bin/python

import socket
import select
import sys
import time
import os
import hashlib

# Prep work
port = int(sys.argv[1])

caching=0

# Check if we might have the -c flag
if len(sys.argv)>3:
    if sys.argv[3].startswith("-c"):
        if len(sys.argv)>4:
            caching = int(sys.argv[4])
        else:
            caching = 3600 # One hour caching by default
    else:
        print "Warning: Did not understand argument "+sys.argv[3]

sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)

sock.bind(("", port))
sock.listen(5)