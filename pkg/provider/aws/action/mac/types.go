package mac

type MacRequestArgs struct {
	// Prefix for the resources related to mac
	// this is relevant in case of an orchestration with multiple
	// macs on the same stack
	Prefix string

	// Machine params
	Architecture string
	Version      string

	// Location params
	FixedLocation    bool
	Region           *string
	AvailabilityZone *string

	// Topology paras
	Airgap bool

	// setup as github actions runner
	SetupGHActionsRunner bool
}

const (
	DefaultArch      = "m2"
	DefaultOSVersion = "15"
)
