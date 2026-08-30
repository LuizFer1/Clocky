//go:build !windows

package termui

import "io"

func enableANSI() {}

func clearNative(w io.Writer) bool {
	_ = w
	return false
}
