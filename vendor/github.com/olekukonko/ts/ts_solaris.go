// +build solaris

// Copyright 2017 Oleku Konko All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// This module is a Terminal  API for the Go Programming Language.
// The protocols were written in pure Go and works on windows and unix systems
package ts

import (
	"syscall"
	"golang.org/x/sys/unix"
)

const (
	TIOCGWINSZ = 21608
)

// Get Windows Size
func GetSize() (ws Size, err error) {
	var wsz *unix.Winsize
	wsz, err = unix.IoctlGetWinsize(syscall.Stdout, TIOCGWINSZ)

	if err != nil {
		ws = Size{80, 25, 0, 0}
	} else {
		ws = Size{wsz.Row, wsz.Col, wsz.Xpixel, wsz.Ypixel}
	}
	return ws, err
}
