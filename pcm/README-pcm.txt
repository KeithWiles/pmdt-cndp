PCM daemon application (Work in progress)

Collects information from the system PCM/perf/PIC/QPI and stores the data
into a shared memory region. The PME tool reads the shared memory region
and displays the data.

Building:
  Make sure you have GNU GCC and G++ installed on your Linux system.
  The version I am using is Ubuntu 19.10 use apt-get to install.

  Make sure you have make installed and execute 'make' in the pcm directory.

  The 'doit' script will execute the daemon/pcm-daemon application using 'sudo' command

You must install the Linux kernel module called 'msr', normally done with this command

sudo modprobe msr

The PCM code is from the https://github.com/opcm/pcm, but modified to exclude all but
the daemon application. The PCM daemon application was also modified to use mmap and
to add more metrics list PCI/QPI data.
