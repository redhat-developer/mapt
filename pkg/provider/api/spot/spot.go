package spot

import (
	"strings"

	mc "github.com/redhat-developer/mapt/pkg/manager/context"
	cr "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
)

type Tolerance int

const (
	Lowest Tolerance = iota
	Low
	Medium
	High
	Highest
)

var (
	tolerances = map[string]Tolerance{
		"lowest":  Lowest,
		"low":     Low,
		"medium":  Medium,
		"high":    High,
		"highest": Highest}

	DefaultTolerance = Lowest

	defaultSpotPriceIncreaseRate = 30
)

func ParseTolerance(str string) (Tolerance, bool) {
	c, ok := tolerances[strings.ToLower(str)]
	return c, ok
}

type SpotArgs struct {
	Spot                  bool
	Tolerance             Tolerance
	IncreaseRate          int
	ExcludedHostingPlaces []string
}

type SpotRequestArgs struct {
	ComputeRequest *cr.ComputeRequestArgs
	OS             *string
	ImageName      *string
	SpotParams     *SpotArgs
}

type SpotResults struct {
	ComputeType      []string
	Price            float64
	HostingPlace     string
	AvailabilityZone string
	ChanceLevel      int
}

type SpotSelector interface {
	Select(mCtx *mc.Context, args *SpotRequestArgs) (*SpotResults, error)
}

// This function add an increased value to the calculated spot price
// to ensure the bid is good enough to have the machine
func SafePrice(basePrice float64, spotPriceIncreseRatio *int) float64 {
	ratio := defaultSpotPriceIncreaseRate
	if spotPriceIncreseRatio != nil {
		ratio = *spotPriceIncreseRatio
	}
	return basePrice * (1 + float64(ratio)/100)
}
