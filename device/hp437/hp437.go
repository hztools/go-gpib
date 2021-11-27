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

package hp437

import (
	"fmt"
	"strconv"
	"strings"

	"hz.tools/gpib"
)

// Device represents a HP 437B to be used over the GPIB / HP-IB.
type Device struct {
	*gpib.Device
}

// Reset will do a soft-reset of the power meter.
func (dev Device) Reset() error {
	_, err := dev.Write([]byte("*RST\r\n"))
	return err
}

// Zero will zero out the sensor against the ref power.
func (dev Device) Zero() error {
	_, err := dev.Write([]byte("ZE\r\n"))
	return err
}

// DisplayUser will display a string to the user. This will only show a few
// chars, so be careful to not blather on too long.
func (dev Device) DisplayUser(s string) error {
	_, err := dev.Write([]byte(fmt.Sprintf("DU%s\r\n", s)))
	return err
}

// Units are the accepted Power reading measures that this device can be
// configured to use.
type Units string

func (u Units) String() string {
	switch u {
	case DBM:
		return "dBm"
	case Watts:
		return "Watts"
	default:
		return string(u)
	}
}

var (
	// Watts can be passed to hp437.Device.Unit() to read power in terms of
	// Watts.
	Watts Units = "LN"

	// DBM can be passed to hp437.Device.Unit() to read power in terms of
	// dBm.
	DBM Units = "LG"
)

// Offset controls for signal loss (e.g., couplers or attenuators).
func (dev Device) Offset(offset float64) error {
	_, err := dev.Write([]byte(fmt.Sprintf("OS%fEN\r\n", offset)))
	return err
}

// Unit will set the Power Meter to read in terms of either Watts or
// dBm.
func (dev Device) Unit(units Units) error {
	_, err := dev.Write([]byte(fmt.Sprintf("%s\r\n", units)))
	return err
}

// Power will return the power (in the configured units) as a floating point
// number.
func (dev Device) Power() (float64, error) {
	buf := make([]byte, 1024)
	i, err := dev.Read(buf)
	if err != nil {
		return 0, err
	}
	buf = buf[:i]

	reading := strings.TrimSpace(string(buf[:i]))
	return strconv.ParseFloat(reading, 64)
}

// New will create a new hp437.Device to use a Power Meter over HP-IP.
func New(dev *gpib.Device) Device {
	return Device{dev}
}

// vim: foldmethod=marker
