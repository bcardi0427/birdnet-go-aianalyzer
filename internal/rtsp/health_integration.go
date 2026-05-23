package rtsp

// This file contains a minimal integration point to wire known-good RTSP bypass
// into the RTSP health check path. The actual wiring into the real health check
// should call ShouldBypassRTSP(url) and short-circuit unhealthy status when true.
// The function below is a placeholder to be used by the health-check caller.

// ApplyBypassIfKnownGood determines whether to bypass health for a given RTSP URL.
// Returns true if bypass should be applied.
func ApplyBypassIfKnownGood(url string) bool {
	return ShouldBypassRTSP(url)
}

// IntegrateBypass is a compatibility wrapper used by the RTSP health checker
// to determine if a given URL should bypass standard health checks.
func IntegrateBypass(url string) bool {
	return ApplyBypassIfKnownGood(url)
}

// IntegrateHealthCheckDecision is a hook that health check callers can use
// to decide whether to bypass the health check for a given RTSP URL.
// It returns true to indicate the health should be bypassed (healthy).
func IntegrateHealthCheckDecision(url string, currentHealthy bool) bool {
	if ApplyBypassIfKnownGood(url) {
		return true
	}
	return currentHealthy
}

// Wire bypass decision into an active health check decision point. If the URL
// is known-good, report healthy regardless of the underlying health probe.
func PatchBypassIntoHealth(url string, currentHealthy bool) bool {
	return IntegrateHealthCheckDecision(url, currentHealthy)
}

// PatchHookHealthCheck applies bypass decision to an RTSP health check result
// and returns the final healthy state. This is a thin wrapper used by callers
// to keep health-check logic centralized.
func PatchHookHealthCheck(url string, currentHealthy bool) bool {
	// If the URL is known-good, bypass health and report healthy
	if ShouldBypassRTSP(url) {
		// Optional: emit a bypass diagnostic log
		// GetLogger().Info("RTSP health bypassed for known-good URL", logger.String("url", url))
		return true
	}
	return currentHealthy
}

// ValidateBypassURL is a small wrapper used by the health checker to confirm
// bypass configuration for a given URL. Returns true if bypass is enabled for this URL.
func ValidateBypassURL(url string) bool {
	return ShouldBypassRTSP(url)
}
