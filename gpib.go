// {{{ Copyright (c) Paul R. Tagliamonte <paul@k3xec.com>, 2021
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE. }}}

package gpib

// TODO(paultag): no pkg-config for gpib yet, so we need to manually set
// the linker using LDFLAGS.

// #cgo LDFLAGS: -lgpib
//
// #include <stdlib.h>
// #include <gpib/ib.h>
import "C"

import (
	"context"
	"fmt"
	"syscall"
	"unsafe"
)

// Options contains configurable aspects of the connected GPIB device.
type Options struct {
	// BaseContext will be used to extend a context with the same lifecycle
	// as the underlying handle to the remote GPIB device.
	BaseContext context.Context

	// Timeout time.Time
	// SendEOI
	// EOS
}

func (opts *Options) context() context.Context {
	if opts == nil || opts.BaseContext == nil {
		return context.Background()
	}
	return opts.BaseContext
}

// Device represents a device connected to the GPIB.
type Device struct {
	ctx        context.Context
	cancel     context.CancelFunc
	closed     bool
	descriptor C.int
}

type status int

func getiberr() error {
	switch C.iberr {
	case 0:
		return syscall.Errno(C.ibcnt)
	case 1:
		return fmt.Errorf("gpib: interface board needs to be controller-in-charge, but is not")
	case 2:
		return fmt.Errorf("gpib: attempted to write data or command bytes, but there are no listeners currently addressed")
	case 3:
		return fmt.Errorf("gpib: interface board has failed to address itself properly before starting an io operation")
	case 4:
		return fmt.Errorf("gpib: arguments to the function call were invalid")
	case 5:
		return fmt.Errorf("gpib: interface board needs to be system controller, but is not")
	case 6:
		return fmt.Errorf("gpib: read or write of data bytes has been aborted")
	case 7:
		return fmt.Errorf("gpib: interface board does not exist")
	case 10:
		return fmt.Errorf("gpib: function call can not proceed due to an asynchronous IO operation")
	case 11:
		return fmt.Errorf("gpib: GPIB board lacks desired capability")
	case 12:
		// filesystem error
		return syscall.Errno(C.ibcnt)
	case 14:
		// TODO(paultag): return a named timeout error here
		return fmt.Errorf("gpib: attempt to write command bytes to the bus has timed out")
	case 15:
		return fmt.Errorf("gpib: serial poll status bytes have been lost")
	case 16:
		return fmt.Errorf("gpib: serial poll request service line is stuck on")
	default:
		return fmt.Errorf("gpib: unknown error")
	}
}

func (s status) Err() error {
	if s&0x8000 == 0x8000 {
		return getiberr()
	}
	return nil
}

// Close will release the underlying handle to the GPIB device, and close
// the related context, terminating any spawned helpers.
func (d *Device) Close() error {
	if d.closed {
		return nil
	}
	d.cancel()
	rv := C.ibonl(d.descriptor, 0)
	if err := status(rv).Err(); err != nil {
		return err
	}
	d.closed = true
	return nil
}

// Local will return local control to the user over the device.
func (d *Device) Local() error {
	rv := C.ibloc(d.descriptor)
	return status(rv).Err()
}

// func (d *Device) Remote(enable bool) error {
// 	var en C.int = 0
// 	if enable {
// 		en = 1
// 	}
// 	rv := C.ibsre(d.descriptor, en)
// 	return status(rv).Err()
// }

// Write will write user data to the GPIB device.
func (d *Device) Write(buf []byte) (int, error) {
	cb := C.CBytes(buf)
	defer C.free(unsafe.Pointer(cb))
	rv := C.ibwrt(d.descriptor, cb, C.long(len(buf)))
	if err := status(rv).Err(); err != nil {
		return 0, err
	}
	return len(buf), nil
}

// Read will read data from the GPIB device.
func (d *Device) Read(buf []byte) (int, error) {
	var (
		cbuflen = C.size_t(len(buf))
		cbuf    = C.malloc(cbuflen)
	)
	// TODO(paultag): Need to check the RV here.
	rv := C.ibrd(d.descriptor, cbuf, C.long(cbuflen))
	if err := status(rv).Err(); err != nil {
		return 0, err
	}

	leng := C.ibcntl
	i := copy(buf, C.GoBytes(cbuf, C.int(leng)))
	return i, nil
}

// Open will open a provided GPIB device.
func Open(board, pad, sad int, opts *Options) (*Device, error) {
	ctx, cancel := context.WithCancel(opts.context())

	desc := C.ibdev(C.int(board), C.int(pad), C.int(sad), 0, 0, 0)
	if desc == -1 {
		return nil, fmt.Errorf("gpib: failed to open the specified device")
	}
	return &Device{
		ctx:        ctx,
		cancel:     cancel,
		descriptor: desc,
	}, nil
}

// vim: foldmethod=marker
