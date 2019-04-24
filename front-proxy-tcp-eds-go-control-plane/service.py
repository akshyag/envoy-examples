#!/usr/bin/env python
import socket
import os

# Create a socket
sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)

# Ensure that you can restart your server quickly when it terminates
sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)

# Set the client socket's TCP "well-known port" number
well_known_port = 8080
sock.bind(('0.0.0.0', well_known_port))

# Set the number of clients waiting for connection that can be queued
sock.listen(5)

# loop waiting for connections (terminate with Ctrl-C)
try:
    while 1:
        newSocket, address = sock.accept(  )
        print "Connected to ", address, "in service ", os.environ['SERVICE_NAME']
        # loop serving the new client
        while 1:
            receivedData = newSocket.recv(1024)
            if not receivedData: break
            # Echo back the same data you just received
            newSocket.send('Hello from behind Envoy (service {})! hostname: {} data got: {}\n'.format(
                os.environ['SERVICE_NAME'],
                socket.gethostname(),
                receivedData))
        newSocket.close(  )
finally:
    sock.close(  )
