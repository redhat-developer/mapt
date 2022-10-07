package spotprice

type SpotPriceRequest struct {
	ProductDescription string
	InstanceType       string
	AvailabilityZones  []string
}

type SpotPriceData struct {
	Price            string
	AvailabilityZone string
	Region           string
	InstanceType     string
}

type SpotPriceResult struct {
	Data SpotPriceData
	Err  error
}
