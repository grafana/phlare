cmake_minimum_required(VERSION 3.22)
project(pyroscope-ebpf C)

set(CMAKE_C_STANDARD 11)

add_library(pg profile.bpf.c)
target_include_directories(pg PRIVATE . vmlinux libbpf)

target_compile_options(pg PRIVATE -nostdinc  -D__TARGET_ARCH_arm64)
