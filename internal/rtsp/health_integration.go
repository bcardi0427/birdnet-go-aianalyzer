package rtsp

// ShouldBypassHealthCheck determines whether standard health checks should be
// bypassed for a given RTSP URL (e.g. if it is a known-good stream).
func ShouldBypassHealthCheck(url string) bool {
	return ShouldBypassRTSP(url)
}

// ApplyBypassToHealthCheck applies the bypass decision to an RTSP health check
// result, returning true (healthy) if bypassed, or the current healthy status otherwise.
func ApplyBypassToHealthCheck(url string, currentHealthy bool) bool {
	if ShouldBypassHealthCheck(url) {
		return true
	}
	return currentHealthy
}

// ApplyBypassIfKnownGood is a legacy wrapper for ShouldBypassHealthCheck.
// Deprecated: Use ShouldBypassHealthCheck instead.
func ApplyBypassIfKnownGood(url string) bool {
	return ShouldBypassHealthCheck(url)
}

// IntegrateBypass is a legacy wrapper for ShouldBypassHealthCheck.
// Deprecated: Use ShouldBypassHealthCheck instead.
func IntegrateBypass(url string) bool {
	return ShouldBypassHealthCheck(url)
}

// IntegrateHealthCheckDecision is a legacy wrapper for ApplyBypassToHealthCheck.
// Deprecated: Use ApplyBypassToHealthCheck instead.
func IntegrateHealthCheckDecision(url string, currentHealthy bool) bool {
	return ApplyBypassToHealthCheck(url, currentHealthy)
}

// PatchBypassIntoHealth is a legacy wrapper for ApplyBypassToHealthCheck.
// Deprecated: Use ApplyBypassToHealthCheck instead.
func PatchBypassIntoHealth(url string, currentHealthy bool) bool {
	return ApplyBypassToHealthCheck(url, currentHealthy)
}

// PatchHookHealthCheck is a legacy wrapper for ApplyBypassToHealthCheck.
// Deprecated: Use ApplyBypassToHealthCheck instead.
func PatchHookHealthCheck(url string, currentHealthy bool) bool {
	return ApplyBypassToHealthCheck(url, currentHealthy)
}

// ValidateBypassURL is a legacy wrapper for ShouldBypassHealthCheck.
// Deprecated: Use ShouldBypassHealthCheck instead.
func ValidateBypassURL(url string) bool {
	return ShouldBypassHealthCheck(url)
}
