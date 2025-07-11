package types

import (
	"strings"

	cr "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
)

const (
	Lowest Tolerance = iota
	Low
	Medium
	High
	Highest

	DefaultTolerance = Lowest
)

type Tolerance int

var tolerances = map[string]Tolerance{
	"lowest":  Lowest,
	"low":     Low,
	"medium":  Medium,
	"high":    High,
	"highest": Highest,
}

func ParseTolerance(str string) (Tolerance, bool) {
	c, ok := tolerances[strings.ToLower(str)]
	return c, ok
}

type SpotRequestArgs struct {
	ComputeRequest *cr.ComputeRequestArgs
	OS             *string
	AMIName        *string
	SpotTolerance  Tolerance
	MaxResults     int
}

type SpotResults struct {
	ComputeType      string
	Price            float32
	Region           string
	AvailabilityZone string
	ChanceLevel      int
}

type SpotSelector interface {
	Select(args *SpotRequestArgs) (*SpotResults, error)
}
