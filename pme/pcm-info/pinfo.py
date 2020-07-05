#! /usr/bin/python3
# SPDX-License-Identifier: BSD-3-Clause
# Copyright(c) 2020 Intel Corporation

import socket
import os
import glob
import json
import readline


CMDS = []


def read_socket(sock, buf_len, echo=True):
    """ Read data from socket and return it in JSON format """
    reply = sock.recv(buf_len).decode()
    try:
        ret = json.loads(reply)
    except json.JSONDecodeError:
        print("Error in reply: ", reply)
        sock.close()
        raise
    if echo:
        print(json.dumps(ret))
    return ret


def handle_socket(path):
    fd = socket.socket(socket.AF_UNIX, socket.SOCK_SEQPACKET)
    global CMDS
    print("Connecting to " + path)
    try:
        fd.connect(path)
    except OSError:
        print("Error connecting to " + path)
        fd.close()
        return
    json_reply = read_socket(fd, 1024)
    output_buf_len = json_reply["max_output_len"]

    fd.send("/".encode())
    CMDS = read_socket(fd, output_buf_len, False)["/"]

    text = input('--> ')
    while (text != "quit"):
        if text.startswith('/'):
            fd.send(text.encode())
            read_socket(fd, output_buf_len)
        text = input('--> ').strip()
    fd.close()

def readline_complete(text, state):
    """ Find any matching commands from the list based on user input """
    all_cmds = ['quit'] + CMDS
    if text:
        matches = [c for c in all_cmds if c.startswith(text)]
    else:
        matches = all_cmds
    return matches[state]


readline.parse_and_bind('tab: complete')
readline.set_completer(readline_complete)
readline.set_completer_delims(readline.get_completer_delims().replace('/', ''))

fd = socket.socket(socket.AF_UNIX, socket.SOCK_SEQPACKET)
# Path to sockets for processes run as a root user
for f in glob.glob('/var/run/pcm-info/pinfo.*'):
  handle_socket(f)
# Path to sockets for processes run as a regular user
for f in glob.glob('/run/user/%d/pcm-info/pinfo.*' % os.getuid()):
  handle_socket(f)
