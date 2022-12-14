..  SPDX-License-Identifier: BSD-3-Clause
    Copyright 2018 The DPDK contributors

.. submitting_patches:

Contributing Code to DPDK
=========================

This document outlines the guidelines for submitting code to PMDT.

The PMDT development process is modeled (loosely) on the Linux Kernel development model so it is worth reading the
Linux kernel guide on submitting patches:
`How to Get Your Change Into the Linux Kernel <https://www.kernel.org/doc/html/latest/process/submitting-patches.html>`_.


The PMDT Development Process
----------------------------

The DPDK development process has the following features:

* The code is hosted in a public git repository.
* There is a mailing list where developers submit patches.
* There are maintainers for hierarchical components.
* Patches are reviewed publicly on the mailing list.
* Successfully reviewed patches are merged to the repository.
* Patches should be sent to the target repository or sub-tree, see below.

The mailing list for DPDK development is `pmdt@dpdk.org <http://mails.dpdk.org/archives/dev/>`_.
Contributors will need to `register for the mailing list <http://mails.dpdk.org/listinfo/dev>`_ in order to submit patches.
It is also worth registering for the DPDK `Patchwork TODO: Insert Patchwork link here`_

The development process requires some familiarity with the ``git`` version control system.
Refer to the `Pro Git Book <http://www.git-scm.com/book/>`_ for further information.

Source License
--------------

Refer to ``LICENSE`` for more details.

Maintainers
-----------

Details are given in the ``CONTRIBUTING.txt`` file.


Getting the Source Code
-----------------------

The source code can be cloned using either of the following:

main repository::

    git clone git://dpdk.org/dpdk
    git clone http://dpdk.org/git/dpdk

sub-repositories (`list <http://git.dpdk.org/next>`_)::

    git clone git://dpdk.org/next/dpdk-next-*
    git clone http://dpdk.org/git/next/dpdk-next-*

Make your Changes
-----------------

Make your planned changes in the cloned ``dpdk`` repo. Here are some guidelines and requirements:

* Follow the :ref:`coding_style` guidelines.

* If you add new files or directories you should add your name to the ``MAINTAINERS`` file.

* Important changes will require an addition to the release notes in ``doc/guides/rel_notes/``.
  See the :ref:`Release Notes section of the Documentation Guidelines <doc_guidelines>` for details.

* Test the compilation works with different targets, compilers and options, see :ref:`contrib_check_compilation`.

* Don't break compilation between commits with forward dependencies in a patchset.
  Each commit should compile on its own to allow for ``git bisect`` and continuous integration testing.

* Add tests to the unit test framework where possible.

* Add documentation, if relevant, in the form of Doxygen comments or a User Guide in RST format.
  See the :ref:`Documentation Guidelines <doc_guidelines>`.

Once the changes have been made you should commit them to your local repo.

For small changes, that do not require specific explanations, it is better to keep things together in the
same patch.
Larger changes that require different explanations should be separated into logical patches in a patchset.
A good way of thinking about whether a patch should be split is to consider whether the change could be
applied without dependencies as a backport.

It is better to keep the related documentation changes in the same patch
file as the code, rather than one big documentation patch at then end of a
patchset. This makes it easier for future maintenance and development of the
code.

As a guide to how patches should be structured run ``git log`` on similar files.


Commit Messages: Subject Line
-----------------------------

The first, summary, line of the git commit message becomes the subject line of the patch email.
Here are some guidelines for the summary line:

* The summary line must capture the area and the impact of the change.

* The summary line should be around 50 characters.

* The summary line should be lowercase apart from acronyms.

* It should be prefixed with the component name (use git log to check existing components).
  For example::

     ixgbe: fix offload config option name

     config: increase max queues per port

* Use the imperative of the verb (like instructions to the code base).

* Don't add a period/full stop to the subject line.

The actual email subject line should be prefixed by ``[PATCH]`` and the version, if greater than v1,
for example: ``PATCH v2``.
The is generally added by ``git send-email`` or ``git format-patch``, see below.

If you are submitting an RFC draft of a feature you can use ``[RFC]`` instead of ``[PATCH]``.
An RFC patch doesn't have to be complete.
It is intended as a way of getting early feedback.


Commit Messages: Body
---------------------

Here are some guidelines for the body of a commit message:

* The body of the message should describe the issue being fixed or the feature being added.
  It is important to provide enough information to allow a reviewer to understand the purpose of the patch.

* When the change is obvious the body can be blank, apart from the signoff.

* The commit message must end with a ``Signed-off-by:`` line which is added using::

      git commit --signoff # or -s

  The purpose of the signoff is explained in the
  `Developer's Certificate of Origin <https://www.kernel.org/doc/html/latest/process/submitting-patches.html#developer-s-certificate-of-origin-1-1>`_
  section of the Linux kernel guidelines.

  .. Note::

     All developers must ensure that they have read and understood the
     Developer's Certificate of Origin section of the documentation prior
     to applying the signoff and submitting a patch.

* The signoff must be a real name and not an alias or nickname.
  More than one signoff is allowed.

* The text of the commit message should be wrapped at 72 characters.

* When fixing a regression, it is required to reference the id of the commit
  which introduced the bug, and put the original author of that commit on CC.
  You can generate the required lines using the following git alias, which prints
  the commit SHA and the author of the original code::

     git config alias.fixline "log -1 --abbrev=12 --format='Fixes: %h (\"%s\")%nCc: %ae'"

  The output of ``git fixline <SHA>`` must then be added to the commit message::

     doc: fix some parameter description

     Update the docs, fixing description of some parameter.

     Fixes: abcdefgh1234 ("doc: add some parameter")
     Cc: author@example.com

     Signed-off-by: Alex Smith <alex.smith@example.com>

* When fixing an error or warning it is useful to add the error message and instructions on how to reproduce it.

* Use correct capitalization, punctuation and spelling.

In addition to the ``Signed-off-by:`` name the commit messages can also have
tags for who reported, suggested, tested and reviewed the patch being
posted. Please refer to the `Tested, Acked and Reviewed by`_ section.

Patch Fix Related Issues
~~~~~~~~~~~~~~~~~~~~~~~~

TODO: Add link for patch review and/or fixing. 


Creating Patches
----------------

It is possible to send patches directly from git but for new contributors it is recommended to generate the
patches with ``git format-patch`` and then when everything looks okay, and the patches have been checked, to
send them with ``git send-email``.

Here are some examples of using ``git format-patch`` to generate patches:

.. code-block:: console

   # Generate a patch from the last commit.
   git format-patch -1

   # Generate a patch from the last 3 commits.
   git format-patch -3

   # Generate the patches in a directory.
   git format-patch -3 -o ~/patch/

   # Add a cover letter to explain a patchset.
   git format-patch -3 -o ~/patch/ --cover-letter

   # Add a prefix with a version number.
   git format-patch -3 -o ~/patch/ -v 2


Cover letters are useful for explaining a patchset and help to generate a logical threading to the patches.
Smaller notes can be put inline in the patch after the ``---`` separator, for example::

   Subject: [PATCH] fm10k/base: add FM10420 device ids

   Add the device ID for Boulder Rapids and Atwood Channel to enable
   drivers to support those devices.

   Signed-off-by: Alex Smith <alex.smith@example.com>
   ---

   ADD NOTES HERE.

    src/pmdt.org/pcm/pcm_structs.go  | 6 ++++++
    src/pmdt.org/pcm/pcm.go | 6 ++++++
    2 files changed, 12 insertions(+)
   ...

Version 2 and later of a patchset should also include a short log of the changes so the reviewer knows what has changed.
This can be added to the cover letter or the annotations.


.. _contrib_checkpatch:

Checking the Patches
--------------------

TODO: Create a method for checking patches for formatting and syntax issues using a script.

Checking Compilation
--------------------

Meson System
~~~~~~~~~~~~

TODO: Create a tool for testing patches with meson build.


Sending Patches
---------------

Patches should be sent to the mailing list using ``git send-email``.
You can configure an external SMTP with something like the following::

   [sendemail]
       smtpuser = name@domain.com
       smtpserver = smtp.domain.com
       smtpserverport = 465
       smtpencryption = ssl

See the `Git send-email <https://git-scm.com/docs/git-send-email>`_ documentation for more details.

The patches should be sent to ``pmdt@dpdk.org``.
If the patches are a change to existing files then you should send them TO the maintainer(s) and CC ``pmdt@dpdk.org``.
The appropriate maintainer can be found in the ``CONTIRBUTING.txt`` file::

   git send-email --to maintainer@some.org --cc pmdt@dpdk.org 000*.patch

New additions can be sent without a maintainer::

   git send-email --to pmdt@dpdk.org 000*.patch

You can test the emails by sending it to yourself or with the ``--dry-run`` option.

If the patch is in relation to a previous email thread you can add it to the same thread using the Message ID::

   git send-email --to pmdt@dpdk.org --in-reply-to <1234-foo@bar.com> 000*.patch

Experienced committers may send patches directly with ``git send-email`` without the ``git format-patch`` step.
The options ``--annotate`` and ``confirm = always`` are recommended for checking patches before sending.


The Review Process
------------------

Patches are reviewed by the community, relying on the experience and
collaboration of the members to double-check each other's work. There are a
number of ways to indicate that you have checked a patch on the mailing list.


Tested, Acked and Reviewed by
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

To indicate that you have interacted with a patch on the mailing list you
should respond to the patch in an email with one of the following tags:

 * Reviewed-by:
 * Acked-by:
 * Tested-by:
 * Reported-by:
 * Suggested-by:

The tag should be on a separate line as follows::

   tag-here: Name Surname <email@address.com>

``Reviewed-by:`` is a strong statement_ that the patch is an appropriate state
for merging without any remaining serious technical issues. Reviews from
community members who are known to understand the subject area and to perform
thorough reviews will increase the likelihood of the patch getting merged.

``Acked-by:`` is a record that the person named was not directly involved in
the preparation of the patch but wishes to signify and record their acceptance
and approval of it.

``Tested-by:`` indicates that the patch has been successfully tested (in some
environment) by the person named.

``Reported-by:`` is used to acknowledge person who found or reported the bug.

``Suggested-by:`` indicates that the patch idea was suggested by the named
person.



Steps to getting your patch merged
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

The more work you put into the previous steps the easier it will be to get a
patch accepted. 

TODO: Add steps for requesting patches to be merged.