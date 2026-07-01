//go:build windows

package nanoarch

import (
	"syscall"
	"unsafe"
)

type dlib struct {
	h syscall.Handle
}

func open(file string) (*dlib, error) {
	h, err := syscall.LoadLibrary(file)
	if err != nil {
		return nil, err
	}
	return &dlib{h: h}, nil
}

func (d *dlib) load(name string) unsafe.Pointer {
	proc, err := syscall.GetProcAddress(d.h, name)
	if err != nil {
		panic("lib function not found: " + name)
	}
	return unsafe.Pointer(proc)
}

func (d *dlib) close() error { return syscall.FreeLibrary(d.h) }
