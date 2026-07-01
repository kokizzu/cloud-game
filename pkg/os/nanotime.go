package os

import _ "unsafe"

//go:linkname nanotime runtime.nanotime
func nanotime() int64

// Nanotime returns a monotonic timestamp in nanoseconds.
func Nanotime() int64 { return nanotime() }
