//
// Created by korniltsev on 7/15/23.
//

#ifndef UMEBPF_UNWIND_FP_ARM64_H
#define UMEBPF_UNWIND_FP_ARM64_H

#include "vmlinux.h"
#include "bpf_helpers.h"


#define PSR_MODE32_BIT        0x00000010
#define PSR_MODE_EL0t    0x00000000
#define PSR_MODE_MASK    0x0000000f

#define compat_user_mode(regs)    \
    (((regs)->pstate & (PSR_MODE32_BIT | PSR_MODE_MASK)) == \
     (PSR_MODE32_BIT | PSR_MODE_EL0t))

struct frame_tail_ {
    u64 fp;
    u64 lr;
};

// backport https://github.com/torvalds/linux/commit/33c222aeda14596ca5b9a1a3002858c6c3565ddd
static __always_inline int unwind_arm64_fp(bpf_user_pt_regs_t *regs, out_stack *stack) {
    int n = 0;
    struct frame_tail_ tail = {};
    u64 fp = regs->regs[29];

    (*stack)[n++] = regs->pc;

    if (compat_user_mode(regs)) {
        return n;
    }

    for (int i = 1; i < UNWIND_MAX_DEPTH; i++) {
        if (fp == 0) {
            break;
        }
        if ((fp & 0x7) != 0) {
            break;
        }
        if (bpf_probe_read_user(&tail, sizeof(tail), (void *) fp)) {
            break;
        }
        (*stack)[n++] = tail.lr; // todo need -1
        if (tail.fp <= fp) {
            break;
        }
        fp = tail.fp;
    }
    return n;
}


#endif //UMEBPF_UNWIND_FP_ARM64_H
