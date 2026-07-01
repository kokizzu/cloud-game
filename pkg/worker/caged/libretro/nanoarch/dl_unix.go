//go:build !windows

package nanoarch

import (
	"fmt"
	"unsafe"
)

/*
#cgo linux LDFLAGS: -ldl
#include <stdlib.h>
#include <dlfcn.h>
*/
import "C"

type dlib struct {
	ptr unsafe.Pointer
}

func open(file string) (*dlib, error) {
	cs := C.CString(file)
	defer C.free(unsafe.Pointer(cs))
	handle := C.dlopen(cs, C.RTLD_LAZY)
	if handle == nil {
		e := C.dlerror()
		if e != nil {
			return nil, fmt.Errorf("couldn't load the lib: %s", C.GoString(e))
		}
		return nil, fmt.Errorf("couldn't load the lib")
	}
	return &dlib{ptr: handle}, nil
}

func (d *dlib) load(name string) unsafe.Pointer {
	cs := C.CString(name)
	defer C.free(unsafe.Pointer(cs))
	ptr := C.dlsym(d.ptr, cs)
	if ptr == nil {
		panic("lib function not found: " + name)
	}
	return ptr
}

func (d *dlib) close() error {
	if d.ptr == nil {
		return nil
	}
	code := int(C.dlclose(d.ptr))
	if code != 0 {
		return fmt.Errorf("couldn't close the lib (%d)", code)
	}
	return nil
}
