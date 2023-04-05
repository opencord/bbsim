#!/bin/bash


# import platform
# platform.processor()
# platform.system        # Windows

# lshw:    width: 64 bits

# >>> import platform
# >>> platform.machine()
# 'x86'


# $ uname -m
# armv7l

## which arch
# https://github.com/hadolint/hadolint/releases/tag/v2.12.0
case "$(uname -a)" in
    *x86_64*)
esac

os=''
case "$(uname -o)" in
    *Linux*) os='Linux'
esac


# hadolint-Darwin-x86_64
# hadolint-Darwin-x86_64.sha256
# hadolint-Linux-arm64
# hadolint-Linux-arm64.sha256
# hadolint-Linux-x86_64
# hadolint-Linux-x86_64.sha256
# hadolint-Windows-x86_64.exe
# hadolint-Windows-x86_64.exe.sha256
# Source code (zip)
# Source code (tar.gz) 
