#ifndef PROFILE_BPF_H
#define PROFILE_BPF_H


#define PERF_MAX_STACK_DEPTH      127
#define PROFILE_MAPS_SIZE         16384

#define SAMPLE_FLAG_USER_STACK_MANUAL 1

struct sample_key {
    __u32 pid;
    __u32 flags;
    __s64 kern_stack;
    __s64 user_stack;
    char comm[16];
};

struct bss_arg {
    __u32 tgid_filter; // 0 => profile everything
    __u8 collect_user;
    __u8 collect_kernel;
};

#define UNWIND_MAX_DEPTH PERF_MAX_STACK_DEPTH
typedef u64 out_stack[UNWIND_MAX_DEPTH];

#endif