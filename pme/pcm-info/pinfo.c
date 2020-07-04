/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright(c) 2020 Intel Corporation
 */

#include <stdio.h>
#include <stdint.h>
#include <unistd.h>
#include <stdlib.h>
#include <stdarg.h>
#include <pthread.h>
#include <sys/socket.h>
#include <sys/un.h>
#include <dlfcn.h>
#include <errno.h>
#include <bsd/string.h>
#include <dirent.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <printf.h>

#include "common.h"
#include "pinfo_private.h"
#include "pinfo.h"

static int list_cmd(pinfo_client_t c);
static int info_cmd(pinfo_client_t c);

static pinfo_t pinfo;

void
pinfo_remove(void)
{
    if (pinfo.sock >= 0) {
        close(pinfo.sock);
		pinfo.sock = -1;
	}

    if (pinfo.sun.sun_path[0]) {
        unlink(pinfo.sun.sun_path);
        pinfo.sun.sun_path[0] = '\0';
    }
}

/*
 * find a command in the list of callbacks
 *
 * @params cmd
 *    The command string to search in the callback list case-insenstive
 * @return
 *    returns the index of the matching command of -1 if not found
 */
static int
find_command(const char *cmd)
{
    int i = -1;

    if (cmd && cmd[0] == '/') {
        for(i = 0; i < pinfo.num_callbacks; i++) {
            if (strcasecmp(cmd, pinfo.callbacks[i].cmd) == 0)
                return i;
        }
    }
    return -1;
}

int
pinfo_register(const char *cmd, pinfo_cb fn)
{
    if (strlen(cmd) >= MAX_CMD_LEN || fn == NULL || cmd[0] != '/')
        return -EINVAL;
    if (pinfo.num_callbacks >= PINFO_MAX_CALLBACKS)
        return -ENOENT;

    if (cmd == NULL || cmd[0] != '/' || fn == NULL)
        return -EINVAL;

    /* Search to see if the command exists already */
    if (find_command(cmd) >= 0)
        return -1;

    strlcpy(pinfo.callbacks[pinfo.num_callbacks].cmd, cmd, MAX_CMD_LEN);
    pinfo.callbacks[pinfo.num_callbacks++].fn = fn;

    return 0;
}

static int
list_cmd(pinfo_client_t _c)
{
    struct pinfo_client *c = _c;
    int i;

    pinfo_append(c, "{%Q:[", c->cmd);
    for (i = 0; i < pinfo.num_callbacks; i++)
        pinfo_append(c, "%Q%s",
            pinfo.callbacks[i].cmd, ((i + 1) < pinfo.num_callbacks)? "," : "");
    pinfo_append(c, "]}");
    return 0;
}

static int
info_cmd(pinfo_client_t _c)
{
    struct pinfo_client *c = _c;
    pinfo_append(c, "{%Q:", c->cmd);
    pinfo_append(c, "{%Q:%d,", "pid", getpid());
    pinfo_append(c, "%Q:%Q,", "version", VERSION);
    pinfo_append(c, "%Q:%d}", "maxbuffer", PINFO_MAX_BUF_LEN);
    pinfo_append(c, "}");
    return 0;
}

static int
invalid_cmd(pinfo_client_t _c)
{
    struct pinfo_client *c = _c;

    pinfo_append(c, "{%Q:", "error");
    if (c->params)
        pinfo_append(c, "\"invalid cmd (%s,%s)\"", c->cmd, c->params);
    else
        pinfo_append(c, "\"invalid cmd (%s)\"", c->cmd);

    pinfo_append(c, "}");
    return 0;
}

static void
perform_command(struct pinfo_client *c, pinfo_cb fn)
{
    int ret = fn(c);

    if (ret < 0) {
        pinfo_append(c, "{null}");
        if (write(c->s, c->buffer, c->used) < 0)
            perror("Error writing to socket");
        return;
    }
    if (write(c->s, c->buffer, c->used) < 0)
        perror("Error writing to socket");

    c->buffer[0] = '\0';
    c->used = 0;
}

static void *
client_handler(void *sock_id)
{
    int bytes, i, s = (int)(uintptr_t)sock_id;
    char info_str[128];
    struct pinfo_client *c;

    snprintf(info_str, sizeof(info_str),
                        "{%Q:%Q,%Q:%d,%Q:%d}", "version", VERSION,
                        "pid", getpid(), "max_output_len", PINFO_MAX_BUF_LEN);
    if (write(s, info_str, strlen(info_str)) < 0) {
        close(s);
        return NULL;
    }

    c = calloc(1, sizeof(struct pinfo_client));
    if (!c)
        return NULL;

    c->s = s;

    for(;;) {
        memset(c->cbuf, 0, MAX_INPUT_CMD_LEN);

        bytes = read(s, c->cbuf, MAX_INPUT_CMD_LEN);
        if (bytes <= 0 || bytes >= MAX_INPUT_CMD_LEN)
            break;

        if (c->cbuf[0] != '/') {
            perform_command(c, invalid_cmd);
            continue;
        }

        c->cmd = strtok_r(c->cbuf, ",", &c->ptr);
        c->params = strtok_r(NULL, ",", &c->ptr);

        i = find_command(c->cmd);
        if (i >= 0)
            perform_command(c, pinfo.callbacks[i].fn);
        else
            perform_command(c, invalid_cmd);
    }
    close(s);

    free(c->buffer);
    free(c);

    return NULL;
}

static void *
socket_listener(void *unused)
{
    (void)unused;

    while (1) {
        pthread_t th;
        int s = accept(pinfo.sock, NULL, NULL);
        if (s < 0) {
            snprintf(pinfo.log_error, sizeof(pinfo.log_error),
                    "Error with accept, process_info thread quitting\n");
            return NULL;
        }
        pthread_create(&th, NULL, client_handler, (void *)(uintptr_t)s);
        pthread_detach(th);
    }
    return NULL;
}

static inline char *
get_socket_path(const char *runtime_dir, const char *prefix)
{
    static char path[1024];

    snprintf(path, sizeof(path), "%s/%s.%d", runtime_dir, prefix, getpid());

    return path;
}

#define DIR_PERMS   0777
static void
_mkdir(const char *dir) {
    char tmp[512];

    snprintf(tmp, sizeof(tmp), "%s", dir);
    size_t len = strlen(tmp);

    if (tmp[len - 1] == '/')
        tmp[len - 1] = 0;

    for(char *p = tmp + 1; *p; p++)
        if (*p == '/') {
            *p = 0;
            mkdir(tmp, DIR_PERMS);
            *p = '/';
        }
    mkdir(tmp, DIR_PERMS);
}

static int
print_quoted(FILE *stream, const struct printf_info *info, const void *const *args)
{
    const char *str = *((const char**)(args[0]));
    return fprintf(stream, "\"%*s\"", (info->left ? -info->width : info->width), str);
}

static int
print_quoted_arginfo(const struct printf_info *info, size_t n, int *argtypes, int *size)
{
    (void)info;
    if (n > 0) {
        argtypes[0] = PA_STRING;
        size[0] = sizeof (char *);
    }
    return 1;
}

int
pinfo_init(const char *runtime_dir, const char *prefix, const char **err_str)
{
    pthread_t t;
    mode_t old;
    const char *dir, *pre;

    register_printf_specifier('Q', print_quoted, print_quoted_arginfo);

    if (!prefix || (strlen(prefix) == 0) || (prefix[0] == '\0'))
        pre = DEFAULT_SOCKET_FILE_PREFIX;

    if (!runtime_dir || (strlen(runtime_dir) == 0) || (runtime_dir[0] == '\0'))
        dir = DEFAULT_SOCKET_LOCATION;

    strlcpy(pinfo.callbacks[pinfo.num_callbacks].cmd, "/", MAX_CMD_LEN);
    pinfo.callbacks[pinfo.num_callbacks++].fn = list_cmd;
    strlcpy(pinfo.callbacks[pinfo.num_callbacks].cmd, "/pcm/info", MAX_CMD_LEN);
    pinfo.callbacks[pinfo.num_callbacks++].fn = info_cmd;

    pinfo.sock = socket(AF_UNIX, SOCK_SEQPACKET, 0);
    if (pinfo.sock < 0) {
        snprintf(pinfo.log_error, sizeof(pinfo.log_error),
                "Error with socket creation, %s", strerror(errno));
        if (err_str)
            *err_str = pinfo.log_error;
        return -1;
    }

    old = umask(0);

	DIR *d = opendir(dir);
	if (d)
		closedir(d);
	else if (errno == ENOENT) {
		_mkdir(dir);
	} else {
		snprintf(pinfo.log_error, sizeof(pinfo.log_error),
			"Error unable to open(%s)", dir);
		goto error;
	}

    pinfo.sun.sun_family = AF_UNIX;
    if (strlcpy(pinfo.sun.sun_path, get_socket_path(dir, pre),
            sizeof(pinfo.sun.sun_path)) >= sizeof(pinfo.sun.sun_path)) {
        snprintf(pinfo.log_error, sizeof(pinfo.log_error),
                "Error with socket binding, path too long");
        goto error;
    }
    if (bind(pinfo.sock, (void *)&pinfo.sun, sizeof(pinfo.sun)) < 0) {
        snprintf(pinfo.log_error, sizeof(pinfo.log_error),
                "Error binding socket (%s): %s", pinfo.sun.sun_path, strerror(errno));
        pinfo.sun.sun_path[0] = 0;
        goto error;
    }

    if (listen(pinfo.sock, 1) < 0) {
        snprintf(pinfo.log_error, sizeof(pinfo.log_error),
                "Error calling listen for socket: %s", strerror(errno));
        goto error;
    }
    umask(old);

    pthread_create(&t, NULL, socket_listener, NULL);
    atexit(pinfo_remove);

    return 0;

error:
	pinfo_remove();
    if (err_str)
        *err_str = pinfo.log_error;
    umask(old);
    return -1;
}

int
pinfo_append(pinfo_client_t _c, const char *format, ...)
{
    struct pinfo_client *c = _c;
    va_list ap;
    char str[1024];
    int ret, nbytes;

    va_start(ap, format);
    ret = vsnprintf(str, sizeof(str), format, ap);
    va_end(ap);

    nbytes = (ret + c->used) + PINFO_EXTRA_SPACE;

    /* Increase size of buffer if required */
    if (nbytes > c->buf_len) {

        /* Make sure the max length is capped to a max size */
        if (nbytes > PINFO_MAX_BUF_LEN)
            return -1;

        /* expand the buffer space */
        char *p = realloc(c->buffer, nbytes);

        if (p == NULL)
            return -1;

        c->buffer = p;
        c->buf_len = nbytes;
    }

    /* Add the new string data to the buffer */
    c->used = strlcat(c->buffer, str, c->buf_len);

    return 0;
}
