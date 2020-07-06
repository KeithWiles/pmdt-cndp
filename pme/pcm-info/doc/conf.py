# SPDX-License-Identifier: BSD-3-Clause
# Copyright(c) 2010-2015 Intel Corporation

from __future__ import print_function
import subprocess
from docutils import nodes
from distutils.version import LooseVersion
from sphinx import __version__ as sphinx_version
from sphinx.highlighting import PygmentsBridge
from pygments.formatters.latex import LatexFormatter
from os import listdir
from os import environ
from os.path import basename
from os.path import dirname
from os.path import join as path_join

try:
    # Python 2.
    import ConfigParser as configparser
except:
    # Python 3.
    import configparser

try:
    import sphinx_rtd_theme

    html_theme = "sphinx_rtd_theme"
    html_theme_path = [sphinx_rtd_theme.get_html_theme_path()]
except:
    print('Install the sphinx ReadTheDocs theme for improved html documentation '
          'layout: pip install sphinx_rtd_theme')
    pass

project = 'Performance Monitor Development Toolkit' 
# html_logo = '../logo/DPDK_logo_vertical_rev_small.png'
# latex_logo = '../logo/DPDK_logo_horizontal_tag.png'
html_add_permalinks = ""
html_show_copyright = False
highlight_language = 'none'

# If MAKEFLAGS is exported by the user, garbage text might end up in version
# version = subprocess.check_output(['make', '-sRrC', '../../', 'showversion'],
#                                  env=dict(environ, MAKEFLAGS=""))
# version = version.decode('utf-8').rstrip()
# release = version

master_doc = 'index'

# Maximum feature description string length
feature_str_len = 25

# Figures, tables and code-blocks automatically numbered if they have caption
numfig = True

latex_documents = [
    ('index',
     'doc.tex',
     '',
     '',
     'manual')
]

# Latex directives to be included directly in the latex/pdf docs.
custom_latex_preamble = r"""
\usepackage{textalpha}
\RecustomVerbatimEnvironment{Verbatim}{Verbatim}{xleftmargin=5mm}
\usepackage{etoolbox}
\robustify\(
\robustify\)
"""

# Configuration for the latex/pdf docs.
latex_elements = {
    'papersize': 'a4paper',
    'pointsize': '11pt',
    # remove blank pages
    'classoptions': ',openany,oneside',
    'babel': '\\usepackage[english]{babel}',
    # customize Latex formatting
    'preamble': custom_latex_preamble
}


# Override the default Latex formatter in order to modify the
# code/verbatim blocks.
class CustomLatexFormatter(LatexFormatter):
    def __init__(self, **options):
        super(CustomLatexFormatter, self).__init__(**options)
        # Use the second smallest font size for code/verbatim blocks.
        self.verboptions = r'formatcom=\footnotesize'

# Replace the default latex formatter.
PygmentsBridge.latex_formatter = CustomLatexFormatter

# Configuration for man pages
man_pages = [("pcm-info/pinfo.py", "pinfo",
              "access pcm-info port stats and memory info", "", 1),
             ("pcm-info/run-pcm", "pcm-info",
              "start the PCM-info daemon", "", 1,),
              ("pme_run", "PMDT",
              "run the performance monitor application written in Go", "", 1)]


# ####### :numref: fallback ########
# The following hook functions add some simple handling for the :numref:
# directive for Sphinx versions prior to 1.3.1. The functions replace the
# :numref: reference with a link to the target (for all Sphinx doc types).
# It doesn't try to label figures/tables.
def numref_role(reftype, rawtext, text, lineno, inliner):
    """
    Add a Sphinx role to handle numref references. Note, we can't convert
    the link here because the doctree isn't build and the target information
    isn't available.
    """
    # Add an identifier to distinguish numref from other references.
    newnode = nodes.reference('',
                              '',
                              refuri='_local_numref_#%s' % text,
                              internal=True)
    return [newnode], []


def process_numref(app, doctree, from_docname):
    """
    Process the numref nodes once the doctree has been built and prior to
    writing the files. The processing involves replacing the numref with a
    link plus text to indicate if it is a Figure or Table link.
    """

    # Iterate over the reference nodes in the doctree.
    for node in doctree.traverse(nodes.reference):
        target = node.get('refuri', '')

        # Look for numref nodes.
        if target.startswith('_local_numref_#'):
            target = target.replace('_local_numref_#', '')

            # Get the target label and link information from the Sphinx env.
            data = app.builder.env.domains['std'].data
            docname, label, _ = data['labels'].get(target, ('', '', ''))
            relative_url = app.builder.get_relative_uri(from_docname, docname)

            # Add a text label to the link.
            if target.startswith('figure'):
                caption = 'Figure'
            elif target.startswith('table'):
                caption = 'Table'
            else:
                caption = 'Link'

            # New reference node with the updated link information.
            newnode = nodes.reference('',
                                      caption,
                                      refuri='%s#%s' % (relative_url, label),
                                      internal=True)
            node.replace_self(newnode)

