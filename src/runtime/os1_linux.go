// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import "unsafe"

var sigset_all sigset = sigset(^uint64(0))

// Linux futex.
//
//	futexsleep(uint32 *addr, uint32 val)
//	futexwakeup(uint32 *addr)
//
// Futexsleep atomically checks if *addr == val and if so, sleeps on addr.
// Futexwakeup wakes up threads sleeping on addr.
// Futexsleep is allowed to wake up spuriously.

const (
	_FUTEX_WAIT = 0
	_FUTEX_WAKE = 1
)

// Atomically,
//	if(*addr == val) sleep
// Might be woken up spuriously; that's allowed.
// Don't sleep longer than ns; ns < 0 means forever.
//go:nosplit
func futexsleep(addr *uint32, val uint32, ns int64) {
	var ts timespec

	// Some Linux kernels have a bug where futex of
	// FUTEX_WAIT returns an internal error code
	// as an errno.  Libpthread ignores the return value
	// here, and so can we: as it says a few lines up,
	// spurious wakeups are allowed.
	if ns < 0 {
		futex(unsafe.Pointer(addr), _FUTEX_WAIT, val, nil, nil, 0)
		return
	}

	// It's difficult to live within the no-split stack limits here.
	// On ARM and 386, a 64-bit divide invokes a general software routine
	// that needs more stack than we can afford. So we use timediv instead.
	// But on real 64-bit systems, where words are larger but the stack limit
	// is not, even timediv is too heavy, and we really need to use just an
	// ordinary machine instruction.
	if ptrSize == 8 {
		ts.set_sec(ns / 1000000000)
		ts.set_nsec(int32(ns % 1000000000))
	} else {
		ts.tv_nsec = 0
		ts.set_sec(int64(timediv(ns, 1000000000, (*int32)(unsafe.Pointer(&ts.tv_nsec)))))
	}
	futex(unsafe.Pointer(addr), _FUTEX_WAIT, val, unsafe.Pointer(&ts), nil, 0)
}

// If any procs are sleeping on addr, wake up at most cnt.
//go:nosplit
func futexwakeup(addr *uint32, cnt uint32) {
	ret := futex(unsafe.Pointer(addr), _FUTEX_WAKE, cnt, nil, nil, 0)
	if ret >= 0 {
		return
	}

	// I don't know that futex wakeup can return
	// EAGAIN or EINTR, but if it does, it would be
	// safe to loop and call futex again.
	systemstack(func() {
		print("futexwakeup addr=", addr, " returned ", ret, "\n")
	})

	*(*int32)(unsafe.Pointer(uintptr(0x1006))) = 0x1006
}

func getproccount() int32 {
	// This buffer is huge (8 kB) but we are on the system stack
	// and there should be plenty of space (64 kB) -- except on ARM where
	// the system stack itself is only 8kb (see golang.org/issue/11873).
	// Also this is a leaf, so we're not holding up the memory for long.
	// See golang.org/issue/11823.
	// The suggested behavior here is to keep trying with ever-larger
	// buffers, but we don't have a dynamic memory allocator at the
	// moment, so that's a bit tricky and seems like overkill.
	const maxCPUs = 64*1024*(1-goarch_arm) + 1024*goarch_arm
	var buf [maxCPUs / (ptrSize * 8)]uintptr
	r := sched_getaffinity(0, unsafe.Sizeof(buf), &buf[0])
	if r < 0 {
		return 1
	}
	n := int32(0)
	for _, v := range buf[:r/ptrSize] {
		for v != 0 {
			n += int32(v & 1)
			v >>= 1
		}
	}
	if n == 0 {
		n = 1
	}
	return n
}

// Clone, the Linux rfork.
const (
	_CLONE_VM             = 0x100
	_CLONE_FS             = 0x200
	_CLONE_FILES          = 0x400
	_CLONE_SIGHAND        = 0x800
	_CLONE_PTRACE         = 0x2000
	_CLONE_VFORK          = 0x4000
	_CLONE_PARENT         = 0x8000
	_CLONE_THREAD         = 0x10000
	_CLONE_NEWNS          = 0x20000
	_CLONE_SYSVSEM        = 0x40000
	_CLONE_SETTLS         = 0x80000
	_CLONE_PARENT_SETTID  = 0x100000
	_CLONE_CHILD_CLEARTID = 0x200000
	_CLONE_UNTRACED       = 0x800000
	_CLONE_CHILD_SETTID   = 0x1000000
	_CLONE_STOPPED        = 0x2000000
	_CLONE_NEWUTS         = 0x4000000
	_CLONE_NEWIPC         = 0x8000000

	cloneFlags = _CLONE_VM | /* share memory */
		_CLONE_FS | /* share cwd, etc */
		_CLONE_FILES | /* share fd table */
		_CLONE_SIGHAND | /* share sig handler table */
		_CLONE_THREAD /* revisit - okay for now */
)

// May run with m.p==nil, so write barriers are not allowed.
//go:nowritebarrier
func newosproc(mp *m, stk unsafe.Pointer) {
	/*
	 * note: strace gets confused if we use CLONE_PTRACE here.
	 */
	mp.tls[0] = uintptr(mp.id) // so 386 asm can find it
	if false {
		print("newosproc stk=", stk, " m=", mp, " g=", mp.g0, " clone=", funcPC(clone), " id=", mp.id, "/", mp.tls[0], " ostk=", &mp, "\n")
	}

	// Disable signals during clone, so that the new thread starts
	// with signals disabled.  It will enable them in minit.
	var oset sigset
	rtsigprocmask(_SIG_SETMASK, &sigset_all, &oset, int32(unsafe.Sizeof(oset)))
	ret := clone(cloneFlags, stk, unsafe.Pointer(mp), unsafe.Pointer(mp.g0), unsafe.Pointer(funcPC(mstart)))
	rtsigprocmask(_SIG_SETMASK, &oset, nil, int32(unsafe.Sizeof(oset)))

	if ret < 0 {
		print("runtime: failed to create new OS thread (have ", mcount(), " already; errno=", -ret, ")\n")
		throw("newosproc")
	}
}

// Version of newosproc that doesn't require a valid G.
//go:nosplit
func newosproc0(stacksize uintptr, fn unsafe.Pointer) {
	stack := sysAlloc(stacksize, &memstats.stacks_sys)
	if stack == nil {
		write(2, unsafe.Pointer(&failallocatestack[0]), int32(len(failallocatestack)))
		exit(1)
	}
	ret := clone(cloneFlags, unsafe.Pointer(uintptr(stack)+stacksize), nil, nil, fn)
	if ret < 0 {
		write(2, unsafe.Pointer(&failthreadcreate[0]), int32(len(failthreadcreate)))
		exit(1)
	}
}

var failallocatestack = []byte("runtime: failed to allocate stack for the new OS thread\n")
var failthreadcreate = []byte("runtime: failed to create new OS thread\n")

func osinit() {
	ncpu = getproccount()
}

var urandom_dev = []byte("/dev/urandom\x00")

func getRandomData(r []byte) {
	if startupRandomData != nil {
		n := copy(r, startupRandomData)
		extendRandom(r, n)
		return
	}
	fd := open(&urandom_dev[0], 0 /* O_RDONLY */, 0)
	n := read(fd, unsafe.Pointer(&r[0]), int32(len(r)))
	closefd(fd)
	extendRandom(r, int(n))
}

func goenvs() {
	goenvs_unix()
}

// Called to initialize a new m (including the bootstrap m).
// Called on the parent thread (main thread in case of bootstrap), can allocate memory.
func mpreinit(mp *m) {
	mp.gsignal = malg(32 * 1024) // Linux wants >= 2K
	mp.gsignal.m = mp
}

//go:nosplit
func msigsave(mp *m) {
	smask := (*sigset)(unsafe.Pointer(&mp.sigmask))
	if unsafe.Sizeof(*smask) > unsafe.Sizeof(mp.sigmask) {
		throw("insufficient storage for signal mask")
	}
	rtsigprocmask(_SIG_SETMASK, nil, smask, int32(unsafe.Sizeof(*smask)))
}

//go:nosplit
func msigrestore(mp *m) {
	smask := (*sigset)(unsafe.Pointer(&mp.sigmask))
	rtsigprocmask(_SIG_SETMASK, smask, nil, int32(unsafe.Sizeof(*smask)))
}

//go:nosplit
func sigblock() {
	rtsigprocmask(_SIG_SETMASK, &sigset_all, nil, int32(unsafe.Sizeof(sigset_all)))
}

func gettid() uint32

// Called to initialize a new m (including the bootstrap m).
// Called on the new thread, can not allocate memory.
func minit() {
	// Initialize signal handling.
	_g_ := getg()
	signalstack(&_g_.m.gsignal.stack)

	// for debuggers, in case cgo created the thread
	_g_.m.procid = uint64(gettid())

	// restore signal mask from m.sigmask and unblock essential signals
	nmask := *(*sigset)(unsafe.Pointer(&_g_.m.sigmask))
	for i := range sigtable {
		if sigtable[i].flags&_SigUnblock != 0 {
			nmask &^= 1 << (uint(i) - 1)
		}
	}
	rtsigprocmask(_SIG_SETMASK, &nmask, nil, int32(unsafe.Sizeof(nmask)))
}

// Called from dropm to undo the effect of an minit.
//go:nosplit
func unminit() {
	signalstack(nil)
}

func memlimit() uintptr {
	/*
		TODO: Convert to Go when something actually uses the result.

		Rlimit rl;
		extern byte runtime·text[], runtime·end[];
		uintptr used;

		if(runtime·getrlimit(RLIMIT_AS, &rl) != 0)
			return 0;
		if(rl.rlim_cur >= 0x7fffffff)
			return 0;

		// Estimate our VM footprint excluding the heap.
		// Not an exact science: use size of binary plus
		// some room for thread stacks.
		used = runtime·end - runtime·text + (64<<20);
		if(used >= rl.rlim_cur)
			return 0;

		// If there's not at least 16 MB left, we're probably
		// not going to be able to do much.  Treat as no limit.
		rl.rlim_cur -= used;
		if(rl.rlim_cur < (16<<20))
			return 0;

		return rl.rlim_cur - used;
	*/

	return 0
}

//#ifdef GOARCH_386
//#define sa_handler k_sa_handler
//#endif

func sigreturn()
func sigtramp()

func setsig(i int32, fn uintptr, restart bool) {
	var sa sigactiont
	memclr(unsafe.Pointer(&sa), unsafe.Sizeof(sa))
	sa.sa_flags = _SA_SIGINFO | _SA_ONSTACK | _SA_RESTORER
	if restart {
		sa.sa_flags |= _SA_RESTART
	}
	sa.sa_mask = ^uint64(0)
	// Although Linux manpage says "sa_restorer element is obsolete and
	// should not be used". x86_64 kernel requires it. Only use it on
	// x86.
	if GOARCH == "386" || GOARCH == "amd64" {
		sa.sa_restorer = funcPC(sigreturn)
	}
	if fn == funcPC(sighandler) {
		fn = funcPC(sigtramp)
	}
	sa.sa_handler = fn
	// Qemu rejects rt_sigaction of SIGRTMAX (64).
	if rt_sigaction(uintptr(i), &sa, nil, unsafe.Sizeof(sa.sa_mask)) != 0 && i != 64 {
		throw("rt_sigaction failure")
	}
}

func setsigstack(i int32) {
	var sa sigactiont
	if rt_sigaction(uintptr(i), nil, &sa, unsafe.Sizeof(sa.sa_mask)) != 0 {
		throw("rt_sigaction failure")
	}
	if sa.sa_handler == 0 || sa.sa_handler == _SIG_DFL || sa.sa_handler == _SIG_IGN || sa.sa_flags&_SA_ONSTACK != 0 {
		return
	}
	sa.sa_flags |= _SA_ONSTACK
	if rt_sigaction(uintptr(i), &sa, nil, unsafe.Sizeof(sa.sa_mask)) != 0 {
		throw("rt_sigaction failure")
	}
}

func getsig(i int32) uintptr {
	var sa sigactiont

	memclr(unsafe.Pointer(&sa), unsafe.Sizeof(sa))
	if rt_sigaction(uintptr(i), nil, &sa, unsafe.Sizeof(sa.sa_mask)) != 0 {
		throw("rt_sigaction read failure")
	}
	if sa.sa_handler == funcPC(sigtramp) {
		return funcPC(sighandler)
	}
	return sa.sa_handler
}

//go:nosplit
func signalstack(s *stack) {
	var st sigaltstackt
	if s == nil {
		st.ss_flags = _SS_DISABLE
	} else {
		st.ss_sp = (*byte)(unsafe.Pointer(s.lo))
		st.ss_size = s.hi - s.lo
		st.ss_flags = 0
	}
	sigaltstack(&st, nil)
}

func updatesigmask(m sigmask) {
	mask := sigset(m)
	rtsigprocmask(_SIG_SETMASK, &mask, nil, int32(unsafe.Sizeof(mask)))
}

func unblocksig(sig int32) {
	if sig > 64 {
		throw("signal > 64")
	}
	mask := sigset(1 << ((uint(sig) - 1)))
	rtsigprocmask(_SIG_UNBLOCK, &mask, nil, int32(unsafe.Sizeof(mask)))
}
