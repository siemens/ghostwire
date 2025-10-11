// IMPORTANT: the //go:build instruction MUST NOT contain a space between the
// comment and the "go:build" instruction; if your beautifier messes this up,
// you'll need to clean it up.

//go:build ignore

#include "common.h"

#define MAX_QUEUE_IDS 128

char __license[] SEC("license") = "Dual MIT/GPL";

/*
 * struct xsksmap maps queue IDs (indices) to their assigned AF_XDP sockets
 * ("xsk"s) if any.
 */
struct {
	__uint(type, BPF_MAP_TYPE_XSKMAP);
	__type(key, __u32);
	__type(value, __u32);
	__uint(max_entries, MAX_QUEUE_IDS);
} xsks_map SEC(".maps");

SEC("xdp")
int xsk_dispatcher(struct xdp_md *ctx) {
	return bpf_redirect_map(&xsks_map, ctx->rx_queue_index, XDP_PASS);
}

/* For testing purposes only: drops all incoming packets. */
SEC("xdp")
int xsk_dropper(struct xdp_md *ctx) {
	return XDP_DROP;
}
