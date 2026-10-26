package telemetry

// The token below is PostHog's project capture token (phc_). It is
// intentionally committed in plain text: it is write-only, can only send
// events into this project, and PostHog documents it as safe to expose in
// client-side code. It must never be replaced with a personal API key
// (phx_), which can read and modify project data.
const (
	posthogHost  = "https://eu.i.posthog.com"
	posthogToken = "phc_BX33W4L8vxDghW9RYfoTn7mTauApz53sDnERYZpUCMQf"
	capturePath  = "/i/v0/e/"
)
