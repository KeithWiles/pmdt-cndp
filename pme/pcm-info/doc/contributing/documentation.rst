..  SPDX-License-Identifier: BSD-3-Clause
    Copyright 2018 The DPDK contributors

.. _doc_guidelines:

PMDT Documentation Guidelines
=============================

This document outlines the guidelines for writing the PMDT documentation in RST format.

It also explains the structure of the PMDT documentation and shows how to build the Html and PDF versions of the documents.


Structure of the Documentation
------------------------------

The PMDT source code repository contains input files to build the documentation.

The main directories that contain files related to documentation are shown below::

   README-pme.txt
   setup.txt
   ...
   pme
   |-- pcm-info
       +-- doc
           |-- contributing
           |-- faq
           |-- howto
           |-- ...


The user guides such as *The Programmers Guide* are generated from RST markup
text files using the `Sphinx <http://sphinx-doc.org>`_ Documentation Generator.

These files are included in the ``pcm-info/doc/`` directory.
The output is controlled by the ``pcm-info/doc/conf.py`` file.


Role of the Documentation
-------------------------

The following items outline the roles of the different parts of the documentation and when they need to be updated or
added to by the developer.

* **Release Notes**

  The Release Notes document which features have been added in the current and previous releases of DPDK and highlight
  any known issues.
  The Releases Notes also contain notifications of features that will change ABI compatibility in the next release.

  Developers should include updates to the Release Notes with patch sets that relate to any of the following sections:

  * New Features
  * Resolved Issues (see below)
  * Known Issues
  * API Changes
  * ABI Changes
  * Shared Library Versions

  Resolved Issues should only include issues from previous releases that have been resolved in the current release.
  Issues that are introduced and then fixed within a release cycle do not have to be included here.

  Refer to the Release Notes from the previous DPDK release for the correct format of each section.

* **The Programmers Guide**

  The Programmers Guide explains how the API components of DPDK such as the EAL, Memzone, Rings and the Hash Library work.
  It also explains how some higher level functionality such as Packet Distributor, Packet Framework and KNI work.
  It also shows the build system and explains how to add applications.

  The Programmers Guide should be expanded when new functionality is added to DPDK.

* **Guidelines**

  The guideline documents record community process, expectations and design directions.

  They can be extended, amended or discussed by submitting a patch and getting community approval.


Building the Documentation
--------------------------

Dependencies
~~~~~~~~~~~~

The following dependencies must be installed to build the documentation:

* Sphinx (also called python-sphinx or python3-sphinx).

* TexLive (at least TexLive-core and the extra Latex support).

* Inkscape.

`Sphinx`_ is a Python documentation tool for converting RST files to Html or to PDF (via LaTeX).
For full support with figure and table captioning the latest version of Sphinx can be installed as follows:

.. code-block:: console

   # Ubuntu/Debian.
   sudo apt-get -y install python-pip
   sudo pip install --upgrade sphinx
   sudo pip install --upgrade sphinx_rtd_theme

   # Red Hat/Fedora.
   sudo dnf     -y install python-pip
   sudo pip install --upgrade sphinx
   sudo pip install --upgrade sphinx_rtd_theme

For further information on getting started with Sphinx see the
`Sphinx Getting Started <http://www.sphinx-doc.org/en/master/usage/quickstart.html>`_.

.. Note::

   To get full support for Figure and Table numbering it is best to install Sphinx 1.3.1 or later.


`Inkscape`_ is a vector based graphics program which is used to create SVG images and also to convert SVG images to PDF images.
It can be installed as follows:

.. code-block:: console

   # Ubuntu/Debian.
   sudo apt-get -y install inkscape

   # Red Hat/Fedora.
   sudo dnf     -y install inkscape


Build commands
~~~~~~~~~~~~~~

The documentation is built using the standard PMDT build system using meson.
Some examples are shown below:

* Generate all the documentation targets::

     In meson_options.txt, set the value of 'enable_docs' to true:
     option('enable_docs', type: 'boolean', value: false,

     Then run: 
     ninja 

The output of the command is generated in the ``build`` directory::

   build/doc
         |-- html
         |   |-- contributing
         |   |-- faq
         |   |.. 


.. Note::

   Make sure to fix any Sphinx warnings when adding or updating documentation.

TODO: Command to remove documentation output files


Document Guidelines
-------------------

Here are some guidelines in relation to the style of the documentation:

* Document the obvious as well as the obscure since it won't always be obvious to the reader.

* Use American English spellings throughout.
  This can be checked using the ``aspell`` utility::

       aspell --lang=en_US --check doc/contributing/documentation.rst


RST Guidelines
--------------

The RST (reStructuredText) format is a plain text markup format that can be converted to Html, PDF or other formats.
It is most closely associated with Python but it can be used to document any language.
It is used in PMDT to document everything.

The Sphinx documentation contains a very useful `RST Primer <http://sphinx-doc.org/rest.html#rst-primer>`_ which is a
good place to learn the minimal set of syntax required to format a document.

The official `reStructuredText <http://docutils.sourceforge.net/rst.html>`_ website contains the specification for the
RST format and also examples of how to use it.
However, for most developers the RST Primer is a better resource.

The most common guidelines for writing RST text are detailed in the
`Documenting Python <https://docs.python.org/devguide/documenting.html>`_ guidelines.
The additional guidelines below reiterate or expand upon those guidelines.


Line Length
~~~~~~~~~~~

* Lines in sentences should be less than 80 characters and wrapped at
  words. Multiple sentences which are not separated by a blank line are joined
  automatically into paragraphs.

* Lines in literal blocks **must** be less than 80 characters since
  they are not wrapped by the document formatters and can exceed the page width
  in PDF documents.


Whitespace
~~~~~~~~~~

* Standard RST indentation is 3 spaces.
  Code can be indented 4 spaces, especially if it is copied from source files.

* No tabs.
  Convert tabs in embedded code to 4 or 8 spaces.

* No trailing whitespace.

* Add 2 blank lines before each section header.

* Add 1 blank line after each section header.

* Add 1 blank line between each line of a list.


Section Headers
~~~~~~~~~~~~~~~

* Section headers should use the following underline formats::

   Level 1 Heading
   ===============


   Level 2 Heading
   ---------------


   Level 3 Heading
   ~~~~~~~~~~~~~~~


   Level 4 Heading
   ^^^^^^^^^^^^^^^


* Level 4 headings should be used sparingly.

* The underlines should match the length of the text.

* In general, the heading should be less than 80 characters, for conciseness.

* As noted above:

   * Add 2 blank lines before each section header.

   * Add 1 blank line after each section header.


Lists
~~~~~

* Bullet lists should be formatted with a leading ``*`` as follows::

     * Item one.

     * Item two is a long line that is wrapped and then indented to match
       the start of the previous line.

     * One space character between the bullet and the text is preferred.

* Numbered lists can be formatted with a leading number but the preference is to use ``#.`` which will give automatic numbering.
  This is more convenient when adding or removing items::

     #. Item one.

     #. Item two is a long line that is wrapped and then indented to match
        the start of the previous line.

     #. Item three.

* Definition lists can be written with or without a bullet::

     * Item one.

       Some text about item one.

     * Item two.

       Some text about item two.

* All lists, and sub-lists, must be separated from the preceding text by a blank line.
  This is a syntax requirement.

* All list items should be separated by a blank line for readability.


Code and Literal block sections
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

* Inline text that is required to be rendered with a fixed width font should be enclosed in backquotes like this:
  \`\`text\`\`, so that it appears like this: ``text``.

* Fixed width, literal blocks of texts should be indented at least 3 spaces and prefixed with ``::`` like this::

     Here is some fixed width text::

        0x0001 0x0001 0x00FF 0x00FF

* It is also possible to specify an encoding for a literal block using the ``.. code-block::`` directive so that syntax
  highlighting can be applied.
  Examples of supported highlighting are::

     .. code-block:: console
     .. code-block:: c
     .. code-block:: go
     .. code-block:: python
     .. code-block:: diff
     .. code-block:: none

  That can be applied as follows::

      .. code-block:: c

         #include<stdio.h>

         int main() {

            printf("Hello World\n");

            return 0;
         }

  Which would be rendered as:

  .. code-block:: c

      #include<stdio.h>

      int main() {

         printf("Hello World\n");

         return 0;
      }


* The default encoding for a literal block using the simplified ``::``
  directive is ``none``.

* Lines in literal blocks must be less than 80 characters since they can exceed the page width when converted to PDF documentation.
  For long literal lines that exceed that limit try to wrap the text at sensible locations.

* Long lines that cannot be wrapped, such as application output, should be truncated to be less than 80 characters.


.. _links:

Hyperlinks
~~~~~~~~~~

* Links to external websites can be plain URLs.
  The following is rendered as http://dpdk.org::

     http://dpdk.org

* They can contain alternative text.
  The following is rendered as `Check out DPDK <http://dpdk.org>`_::

     `Check out DPDK <http://dpdk.org>`_

* An internal link can be generated by placing labels in the document with the format ``.. _label_name``.

* The following links to the top of this section: :ref:`links`::

     .. _links:

     Hyperlinks
     ~~~~~~~~~~

     * The following links to the top of this section: :ref:`links`:

.. Note::

   The label must have a leading underscore but the reference to it must omit it.
   This is a frequent cause of errors and warnings.

* The use of a label is preferred since it works across files and will still work if the header text changes.

* Read the rendered section of the documentation that you have added for correctness, clarity and consistency
  with the surrounding text.