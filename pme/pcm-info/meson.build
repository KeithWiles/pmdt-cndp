# SPDX-License-Identifier: BSD-3-Clause
# Copyright(c) 2019-2020 Intel Corporation

project('PCM-Info', ['cpp', 'c'],
    # Fallback to "more" for Windows compatibility.
    version: run_command(find_program('cat', 'more'),
        files('../../VERSION')).stdout().strip(),
    license: 'BSD',
    default_options: ['buildtype=release', 'default_library=static'],
    meson_version: '>= 0.47.1'
)

# set up some global vars for compiler, platform, configuration, etc.
cc = meson.get_compiler('c')
#ccpp = meson.get_compiler('cpp')

opcm_dir = get_option('opcm_dir')

add_project_arguments('-g', language: ['c', 'cpp'])
add_project_arguments('-DPCM_USE_PERF', language: ['c', 'cpp'])
add_project_arguments('-I' + opcm_dir, language: ['c', 'cpp'])
add_project_arguments('-Wno-format', language: ['c', 'cpp'])
add_project_arguments('-Wno-format-extra-args', language: ['c', 'cpp'])
add_project_link_arguments('-pthread', language: ['c', 'cpp'])

# check for libbsd
if cc.find_library('libbsd', required: true).found()
    add_project_link_arguments('-lbsd', language: ['c', 'cpp'])
endif

if cc.find_library('rt', required : true).found()
    add_project_link_arguments('-lrt', language: ['c', 'cpp'])
endif

files = ['main.cpp', 'pcm-info.cpp', 'pinfo.c', 'system-info.c']

opcm_files = ['utils', 'pci', 'msr', 'mmio', 'client_bw', 'cpucounters']
foreach f:opcm_files
    files += opcm_dir + '/' + f + '.cpp'
endforeach
foreach f:files
    message('File: ' + f)
endforeach
executable('pcm-info', files)
