/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright(c) 2020 Intel Corporation
 */

#include "pcm-info.h"

int main(int argc, char *argv[])
{
	PCMDaemon::Daemon daemon(argc, argv);

	return daemon.run();
}
