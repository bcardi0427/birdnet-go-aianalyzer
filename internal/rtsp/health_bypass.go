package rtsp

import "os"

// IsKnownGoodRTSP returns true if the given URL should be treated as healthy
// regardless of health probe results. This uses an environment variable flag
// to opt-in for known-good RTSP streams during debugging or staged rollouts.
func IsKnownGoodRTSP(url string) bool {
	v := os.Getenv("BIRDNET_KNOWN_GOOD_RTSP")
	if v == "" {
		return false
	}
	// Simple containment check to avoid brittle URL parsing
	return contains(url, v)
}

func contains(s, sub string) bool {
	return len(sub) > 0 && (stringIndex(s, sub) >= 0)
}

// stringIndex is a tiny wrapper to avoid importing strings in this file.
func stringIndex(s, t string) int {
	for i := 0; i+len(t) <= len(s); i++ {
		if s[i:i+len(t)] == t {
			return i
		}
	}
	return -1
}

// ShouldBypassRTSP is a small helper intended to be called by the RTSP
// health checker to determine if the provided RTSP URL should bypass
// standard health checks due to being known-good.
// It delegates to IsKnownGoodRTSP for the logic and exists to keep the
// health-check integration surface stable.
func ShouldBypassRTSP(url string) bool {
	return IsKnownGoodRTSP(url)
}
