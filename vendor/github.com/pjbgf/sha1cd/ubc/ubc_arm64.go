//go:build !noasm && gc && arm64 && !amd64
// +build !noasm,gc,arm64,!amd64

package ubc

func CalculateDvMaskARM64(W [80]uint32) uint32

// Check takes as input an expanded message block and verifies the unavoidable
// bitconditions for all listed DVs. It returns a dvmask where each bit belonging
// to a DV is set if all unavoidable bitconditions for that DV have been met.
func CalculateDvMask(W [80]uint32) uint32 {
	return CalculateDvMaskARM64(W)
}
