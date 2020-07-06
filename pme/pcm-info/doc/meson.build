# SPDX-License-Identifier: BSD-3-Clause
# Copyright(c) 2018 Intel Corporation

python3 = import('python').find_installation(required: false)
if python3.found()
        py3 = [python3]
else
        py3 = ['meson', 'runpython']
endif
sphinx          = find_program('sphinx-build', required: get_option('enable_docs'))
sphinx_wrapper	= [python3] + files('call-sphinx-build.py')

current_dir 	= meson.source_root()
build_dir	= join_paths(current_dir, 'build-docs')

if not sphinx.found()
        subdir_done()
endif

# htmldir = join_paths(current_dir, 'doc', 'pmdt')
html_guides = custom_target('html_guides',
        input: files('index.rst'),
        output: 'html',
        command: [sphinx_wrapper, sphinx, meson.current_source_dir(), meson.current_build_dir()],
        depfile: '.html.d',
        build_by_default: get_option('enable_docs'),
        install: get_option('enable_docs'),
        install_dir: build_dir)

install_data(files('custom.css'), install_dir: join_paths(build_dir, '_static', 'css'))

message = 'Building docs:'

run_target('doc', command: ['echo', message, 'HTML_Guides'],
	depends: html_guides)

