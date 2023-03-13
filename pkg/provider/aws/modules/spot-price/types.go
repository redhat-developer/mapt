package spotprice

type SpotPriceRequest struct {
	TargetHostID string
	Name         string
}

type SpotPriceResult struct {
	Prices []SpotPriceGroup
	Err    error
}

type SpotPriceGroup struct {
	Region           string
	AvailabilityZone string
	AVGPrice         float64
	MaxPrice         float64
	Score            int64
}
