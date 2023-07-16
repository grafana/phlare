// SPDX-License-Identifier: GPL-2.0-only

#include "vmlinux.h"
#include "bpf_helpers.h"
#include "profile.bpf.h"


struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __type(key, struct sample_key);
    __type(value, u32);
    __uint(max_entries, PROFILE_MAPS_SIZE);
} counts SEC(".maps");


struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __type(key, u32);
    __type(value, out_stack);
    __uint(max_entries, PROFILE_MAPS_SIZE);
} manual_stacks SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __type(key, u32);
    __type(value, out_stack);
    __uint(max_entries, 1);
} scratch_stacks SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_STACK_TRACE);
    __uint(key_size, sizeof(u32));
    __uint(value_size, PERF_MAX_STACK_DEPTH * sizeof(u64));
    __uint(max_entries, PROFILE_MAPS_SIZE);
} stacks SEC(".maps");


struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __type(key, u32);
    __type(value, struct bss_arg);
    __uint(max_entries, 1);
} args SEC(".maps");

struct bss_arg arg2;

#define KERN_STACKID_FLAGS (0 | BPF_F_FAST_STACK_CMP)
#define USER_STACKID_FLAGS (0 | BPF_F_FAST_STACK_CMP | BPF_F_USER_STACK)


#if defined(__TARGET_ARCH_arm64)
//todo we l probably need hash out of this ifdef once we have python and ruby
#include "hash.h"
#include "unwind_fp_arm64.h"

static __always_inline u32 get_arm64_user_stack_hash(bpf_user_pt_regs_t *regs, out_stack *stack) {
    for (int i = 0; i < UNWIND_MAX_DEPTH; i++) { //todo just put a trailing zero in unwind_arm64_fp
        (*stack)[i] = 0;
    }
    int n = unwind_arm64_fp(regs, stack);
    u32 stack_hash = MurmurHash2(stack, n * (int) sizeof(u64), 0);
    long err = bpf_map_update_elem(&manual_stacks, &stack_hash, stack, BPF_ANY);
    if (err != 0) {
        return -1;
    }
    return stack_hash;
}

#endif

SEC("perf_event")
int do_perf_event(struct bpf_perf_event_data *ctx) {
    u64 id = bpf_get_current_pid_tgid();
    u32 tgid = id >> 32;
    u32 pid = id;
    struct sample_key key = {.pid = tgid};
    key.kern_stack = -1;
    key.flags = -1;
    key.user_stack = -1;
    u32 *val, one = 1, zero = 0;
    out_stack *scratch_stack;
    struct bss_arg *arg = bpf_map_lookup_elem(&args, &zero);
    if (!arg) {
        return 0;
    }
    if (pid == 0) {
        return 0;
    }
    if (arg->tgid_filter != 0 && tgid != arg->tgid_filter) {
        return 0;
    }

    bpf_get_current_comm(&key.comm, sizeof(key.comm));

    if (arg->collect_kernel) {
        key.kern_stack = bpf_get_stackid(ctx, &stacks, KERN_STACKID_FLAGS);
    }
    if (arg->collect_user) {

        // todo maybe ifdef kernel version
#if defined(__TARGET_ARCH_arm64)
        key.flags = SAMPLE_FLAG_USER_STACK_MANUAL;
        scratch_stack = bpf_map_lookup_elem(&scratch_stacks, &zero);

        key.user_stack = get_arm64_user_stack_hash(&ctx->regs, scratch_stack);
#else
        key.user_stack = bpf_get_stackid(ctx, &stacks, USER_STACKID_FLAGS);
#endif
    }

    val = bpf_map_lookup_elem(&counts, &key);
    if (val)
        (*val)++;
    else
        bpf_map_update_elem(&counts, &key, &one, BPF_NOEXIST);
    return 0;
}

char _license[] SEC("license") = "GPL";
