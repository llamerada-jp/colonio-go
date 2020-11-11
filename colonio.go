package colonio

import (
	// #cgo CFLAGS: -I${SRCDIR}/../colonio/src
	// #cgo LDFLAGS: -L${SRCDIR}/../colonio/output -L${SRCDIR}/../colonio/output/lib -lcolonio -lwebrtc -lm -lprotobuf -lstdc++
	// #cgo pkg-config: openssl
	// #include <stdlib.h>
	// #include <colonio/colonio.h>
	"C"
	"fmt"
)

// Colonio is an instance. It is equivalent to one node.
type Colonio struct {
	colonioPtr C.struct_colonio_s
}

func convertError(err *C.struct_colonio_error_s) error {
	return fmt.Errorf("colonio error")
}

// NewColonio creates a new initialized instance.
func NewColonio() (*Colonio, error) {
	instance := &Colonio{}
	err := C.colonio_init(&instance.colonioPtr)
	if err != nil {
		return nil, convertError(err)
	}
	return instance, nil
}

// Connect to seed and join the cluster.
func (c *Colonio) Connect(url string, token string) error {
	err := C.colonio_connect(&c.colonioPtr,
		C.CString(url), C.uint(len(url)),
		C.CString(token), C.uint(len(token)))
	if err != nil {
		return convertError(err)
	}
	return nil
}

// Disconnect from the cluster and the seed.
func (c *Colonio) Disconnect() error {
	err := C.colonio_disconnect(&c.colonioPtr)
	if err != nil {
		return convertError(err)
	}
	return nil
}

// Quit is the finalizer of the instance.
func (c *Colonio) Quit() error {
	err := C.colonio_quit(&c.colonioPtr)
	if err != nil {
		return convertError(err)
	}
	return nil
}
