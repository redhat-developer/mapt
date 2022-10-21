package spotprice

type SpotPriceRequest struct {
	ProductDescription string
	InstanceType       string
	AvailabilityZones  []string
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
