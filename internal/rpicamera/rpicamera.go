package rpicamera

import (
	"fmt"
	"unsafe"
)

// #cgo LDFLAGS: -ldl
// #include <stdlib.h>
// #include <dlfcn.h>
import "C"

type RPICamera struct{}

func New() (*RPICamera, error) {
	str := C.CString("/opt/vc/lib/libmmal_components.so")
	defer C.free(unsafe.Pointer(str))

	handle := C.dlopen(str, C.RTLD_LAZY)
	if handle == nil {
		return nil, fmt.Errorf("unable to open MMAL library")
	}
	defer C.dlclose(handle)

	fmt.Println(handle)

	return nil, fmt.Errorf("TODO")
}
