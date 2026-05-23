//go:build cgo

package tflite

import (
	"github.com/tphakala/go-tflite/delegates"
	"github.com/tphakala/go-tflite/delegates/xnnpack"
)

func newXNNPACKDelegate(threads int) delegates.Delegater {
	return xnnpack.New(xnnpack.DelegateOptions{NumThreads: int32(max(1, threads-1))}) //nolint:gosec // G115: thread count bounded by CPU count, safe conversion
}
