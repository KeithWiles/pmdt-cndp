..  SPDX-License-Identifier: BSD-3-Clause
    Copyright(c) 2010-2014 Intel Corporation.

What does "EAL: map_all_hugepages(): open failed: Permission denied Cannot init memory" mean?
---------------------------------------------------------------------------------------------

This is most likely due to the test application not being run with sudo to promote the user to a superuser.
Alternatively, applications can also be run as regular user.
For more information, please refer to :ref:`DPDK Getting Started Guide <linux_gsg>`.


If I want to change the number of hugepages allocated, how do I remove the original pages allocated?
----------------------------------------------------------------------------------------------------

The number of pages allocated can be seen by executing the following command::

   grep Huge /proc/meminfo

Once all the pages are mmapped by an application, they stay that way.
If you start a test application with less than the maximum, then you have free pages.
When you stop and restart the test application, it looks to see if the pages are available in the ``/dev/huge`` directory and mmaps them.
If you look in the directory, you will see ``n`` number of 2M pages files. If you specified 1024, you will see 1024 page files.
These are then placed in memory segments to get contiguous memory.

If you need to change the number of pages, it is easier to first remove the pages. The usertools/dpdk-setup.sh script provides an option to do this.
See the "Quick Start Setup Script" section in the :ref:`DPDK Getting Started Guide <linux_gsg>` for more information.

