#! /usr/bin/python3

import socket
import os
import sys
import time
import glob

fd = socket.socket(socket.AF_UNIX, socket.SOCK_SEQPACKET)
for f in glob.glob('/var/run/pcm-info/pinfo.*'):
    print("Connecting to " + f)
    try:
        fd.connect(f)
    except OSError:
        continue
    text = input('--> ')
    while (text != "quit"):
        fd.send(text.encode())
        d = fd.recv(64*1024)
        print(d.decode())
        text = input('--> ')
    fd.close()
