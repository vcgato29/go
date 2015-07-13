// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build s390x

package runtime

import "unsafe"

// The calls to nop are to keep these functions from being inlined.
// If they are inlined we have no guarantee that later rewrites of the
// code by optimizers will preserve the relative order of memory accesses.

//go:nosplit
func atomicload(ptr *uint32) uint32 {
	nop()
	return *ptr
}

//go:nosplit
func atomicloadp(ptr unsafe.Pointer) unsafe.Pointer {
	nop()
	return *(*unsafe.Pointer)(ptr)
}

//go:nosplit
func atomicload64(ptr *uint64) uint64 {
	nop()
	return *ptr
}

//go:noescape
func xadd(ptr *uint32, delta int32) uint32

//go:noescape
func xadd64(ptr *uint64, delta int64) uint64

//go:noescape
func xadduintptr(ptr *uintptr, delta uintptr) uintptr

//go:noescape
func xchg(ptr *uint32, new uint32) uint32

//go:noescape
func xchg64(ptr *uint64, new uint64) uint64

// NO go:noescape annotation; see atomic_pointer.go.
func xchgp1(ptr unsafe.Pointer, new unsafe.Pointer) unsafe.Pointer

//go:noescape
func xchguintptr(ptr *uintptr, new uintptr) uintptr

func atomicor8(addr *uint8, v uint8) {
	// Align down to 4 bytes and use 32-bit CAS.
	uaddr := uintptr(unsafe.Pointer(addr))
	addr32 := (*uint32)(unsafe.Pointer(uaddr &^ 3))
	word := uint32(v) << (((uaddr & 3) ^ 3) * 8) // big endian
	for {
		old := *addr32
		if cas(addr32, old, old|word) {
			return
		}
	}
}

func atomicand8(addr *uint8, v uint8) {
	// Align down to 4 bytes and use 32-bit CAS.
	uaddr := uintptr(unsafe.Pointer(addr))
	addr32 := (*uint32)(unsafe.Pointer(uaddr &^ 3))
	shift_bits := ((uaddr & 3) ^ 3) * 8  // big endian
	word := uint32(v) << (shift_bits)    // big endian
	mask := uint32(0xFF) << (shift_bits) // big endian
	word |= ^mask
	for {
		old := *addr32
		if cas(addr32, old, old&word) {
			return
		}
	}
}

// NOTE: Do not add atomicxor8 (XOR is not idempotent).

//go:noescape
func cas64(ptr *uint64, old, new uint64) bool

//go:noescape
func atomicstore(ptr *uint32, val uint32)

//go:noescape
func atomicstore64(ptr *uint64, val uint64)

// NO go:noescape annotation; see atomic_pointer.go.
func atomicstorep1(ptr unsafe.Pointer, val unsafe.Pointer)
