/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright(c) 2020 Intel Corporation
 */

#include <stdint.h>

#ifndef _PINFO_H_
#define _PINFO_H_

#ifdef __cplusplus
extern "C" {
#endif

#ifndef _PINFO_PRIVATE_H_
typedef void *pinfo_client_t;

/* callback returns json data in buffer, up to buf_len long.
 * returns length of buffer used on success, negative on error.
 */
typedef int (*pinfo_cb)(pinfo_client_t c);
#endif

int pinfo_register(const char *cmd, pinfo_cb fn);

int pinfo_init(const char *runtime_dir, const char *prefix, const char **err_str);

int pinfo_append(pinfo_client_t client, const char *format, ...);

void pinfo_remove(void);

#ifdef __cplusplus
}
#endif

#endif /* _PINFO_H_ */
