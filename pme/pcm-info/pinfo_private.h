/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright(c) 2020 Intel Corporation
 */

#include <stdint.h>

#ifndef _PINFO_PRIVATE_H_
#define _PINFO_PRIVATE_H_

#define DEFAULT_SOCKET_LOCATION		"/var/run/pcm-info"
#define DEFAULT_SOCKET_FILE_PREFIX	"pcm-data"

#define PINFO_MAX_CALLBACKS 64

#define PINFO_EXTRA_SPACE   64
#define PINFO_MAX_BUF_LEN   (16 * 1024)

#ifndef _PINFO_H_
typedef void *pinfo_client_t;

/* callback returns json data in buffer, up to buf_len long.
 * returns length of buffer used on success, negative on error.
 */
typedef int (*pinfo_cb)(pinfo_client_t c);
#endif

#define MAX_INPUT_CMD_LEN   256
#define MAX_CMD_LEN         56
struct cmd_callback {
    char cmd[MAX_CMD_LEN];
    pinfo_cb fn;
};

typedef struct {
    int sock;
    struct sockaddr_un sun;
    char log_error[1024];
    struct cmd_callback callbacks[PINFO_MAX_CALLBACKS];
    int num_callbacks;
} pinfo_t;

struct pinfo_client {
    char cbuf[MAX_INPUT_CMD_LEN];
    char *buffer;
    char *ptr;
    const char *cmd;
    const char *params;
    int buf_len;
    int used;
    int s;
};

#endif /* _PINFO_PRIVATE_H_ */
