#!/usr/bin/python
import socket
import sys
import base64
import hmac

HOST, PORT = "localhost", 8888

# SOCK_DGRAM is the socket type to use for UDP sockets
sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)

# As you can see, there is no connect() call; UDP has no connections.
# Instead, data is directly sent to the recipient via sendto().
name=sys.argv[1].encode()
code=sys.argv[2].encode()
msg=sys.argv[3].encode()
sock.sendto(name+b"\n"+base64.encodestring(hmac.new(code,msg).digest())[:-1]+b'\0'+msg, (HOST, PORT))
received = sock.recv(1024), "utf-8"

print(repr(received[0]))


