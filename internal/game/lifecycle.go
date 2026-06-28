package game

import "time"

const (
	LifecycleActive        = "active"
	LifecycleWarned        = "warned"
	LifecycleHidden        = "hidden"
	LifecyclePurgeEligible = "purge_eligible"

	LifecycleWarnAfter          = 14 * 24 * time.Hour
	LifecycleHideAfter          = 30 * 24 * time.Hour
	LifecyclePurgeEligibleAfter = 60 * 24 * time.Hour
)
