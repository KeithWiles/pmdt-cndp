..  SPDX-License-Identifier: BSD-3-Clause
    Copyright 2018 The DPDK contributors

.. _coding_style:

PMDT Coding Style
=================

Description
-----------

This document specifies the preferred style for source files in the PMDT source tree.
It is based on the Linux Kernel coding guidelines and the FreeBSD 7.2 Kernel Developer's Manual (see man style(9)) as well as adapted from DPDK coding guidelines.

General Guidelines
------------------

The rules and guidelines given in this document cannot cover every situation, so the following general guidelines should be used as a fallback:

* The code style should be consistent within each individual file.
* In the case of creating new files, the style should be consistent within each file in a given directory or module.
* The primary reason for coding standards is to increase code readability and comprehensibility, therefore always use whatever option will make the code easiest to read.

Line length is recommended to be not more than 80 characters, including comments.
[Tab stop size should be assumed to be 8-characters wide].

.. note::

	The above is recommendation, and not a hard limit.
	However, it is expected that the recommendations should be followed in all but the rarest situations.

C Comment Style
---------------

Usual Comments
~~~~~~~~~~~~~~

These comments should be used in normal cases.

.. code-block:: c

 /*
  * VERY important single-line comments look like this.
  */

 /* Most single-line comments look like this. */

 /*
  * Multi-line comments look like this.  Make them real sentences. Fill
  * them so they look like real paragraphs.
  */

License Header
~~~~~~~~~~~~~~

Each file should begin with a special comment containing the appropriate copyright and license for the file.
Generally this is the BSD License, except for code for Linux Kernel modules.
After any copyright header, a blank line should be left before any other contents, e.g. include statements in a go file.

GO Comment Style
---------------

Usual Comments
~~~~~~~~~~~~~~

These comments should be used in normal cases.

.. code-block:: go

 //
 // VERY important single-line comments look like this.
 //

 // Most single-line comments look like this. 

 // Multi-line comments look like this.  Make them real sentences. Fill
 // them so they look like real paragraphs.
 

License Header
~~~~~~~~~~~~~~

Each file should begin with a special comment containing the appropriate copyright and license for the file.
Generally this is the BSD License, except for code for Linux Kernel modules.
After any copyright header, a blank line should be left before any other contents, e.g. include statements in a C file.
