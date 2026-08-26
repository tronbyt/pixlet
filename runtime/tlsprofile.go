package runtime

import (
	_ "unsafe" // for go:linkname
)

// Go orders the TLS 1.3 cipher suites it offers according to whether the CPU
// has hardware AES: AES-GCM first where the instructions exist, ChaCha20 first
// where they don't, because ChaCha20 is the faster choice in software. The
// decision is made once at package init from CPU feature detection and is not
// reachable through tls.Config — TLS 1.3 suites are explicitly not
// configurable there.
//
// The practical effect is that the same pixlet binary presents two different
// TLS fingerprints depending on the host it runs on, and some origins treat
// the two differently. reddit.com is one: it serves a 403 block page to the
// ChaCha20-first ordering and 200 to the AES-first one. Devices without the
// ARMv8 crypto extensions — the Raspberry Pi 3, Pi 4 and Zero 2 W, and any
// 32-bit Pi OS build — therefore cannot fetch from hosts that x86 and Pi 5
// servers reach without trouble, and an app author has no way to tell or to
// work around it.
//
// Measured against reddit.com from one machine, same address, one minute
// apart, with only this ordering changed:
//
//	4865-4866-4867 (AES first, what x86 sends)     -> 200
//	4867-4865-4866 (ChaCha first, what a Pi sends) -> 403
//
// Aligning the two orderings makes a pixlet host's fingerprint independent of
// its CPU, so an app behaves the same everywhere. Both variables are pulled by
// linkname, which crypto/tls supports for exactly this kind of use
// (go.dev/issue/67401); no linker flags and no extra dependencies are needed.
//
// The cost is that TLS 1.3 connections on a machine without hardware AES will
// now negotiate AES-GCM rather than ChaCha20, which is slower in software.
// pixlet's traffic is a handful of requests per render and is dominated by
// image decoding, so consistent reachability is the better trade.

//go:linkname defaultCipherSuitesTLS13 crypto/tls.defaultCipherSuitesTLS13
var defaultCipherSuitesTLS13 []uint16

//go:linkname defaultCipherSuitesTLS13NoAES crypto/tls.defaultCipherSuitesTLS13NoAES
var defaultCipherSuitesTLS13NoAES []uint16

func init() {
	// Same suites in both, only the order differs, so this is a reordering
	// rather than a change to what is offered or accepted.
	if len(defaultCipherSuitesTLS13NoAES) == len(defaultCipherSuitesTLS13) {
		copy(defaultCipherSuitesTLS13NoAES, defaultCipherSuitesTLS13)
	}
}
