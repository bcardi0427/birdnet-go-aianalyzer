//go:build !cgo

package tflite

import "github.com/tphakala/go-tflite/delegates"

func newXNNPACKDelegate(_ int) delegates.Delegater {
	return nil
}
