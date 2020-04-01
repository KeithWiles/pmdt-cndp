#! /usr/bin/python3
# SPDX-License-Identifier: BSD-3-Clause
# Copyright(c) 2020 Intel Corporation

import socket
import os
import glob
import json

def handle_socket(path):
    print("Connecting to " + path)
    try:
        fd.connect(path)
    except OSError:
        return
    text = input('--> ')
    while (text != "quit"):
        fd.send(text.encode())
        reply = json.loads(fd.recv(1024 * 12).decode())
        print(json.dumps(reply))
        text = input('--> ')
    fd.close()

fd = socket.socket(socket.AF_UNIX, socket.SOCK_SEQPACKET)
# Path to sockets for processes run as a root user
for f in glob.glob('/var/run/pcm-info/pinfo.*'):
  handle_socket(f)
# Path to sockets for processes run as a regular user
for f in glob.glob('/run/user/%d/pcm-info/pinfo.*' % os.getuid()):
  handle_socket(f)
