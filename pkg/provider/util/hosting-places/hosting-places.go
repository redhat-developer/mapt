package hostingplace

import "sync"

// Conceptually each cloud provider works in a similar way
// they offers services across different zones:
//
// * AWS those are called Regions
// * Azure Locations
// * GCP Zones
//
// We wil name the concept with as hostingplace and this class will help
// for those operations which are required to be executed in parallel across
// several or all of them per provider

// Struct to communicate data tied region
// when running some aggregation data func async on a number of regions

type HostingPlaceData[Y any] struct {
	Region string
	Err    error
	Value  Y
}

// Generic function to run specific function on each region
// and then aggregate the results into a struct
func RunOnHostingPlaces[X, Y any](hps []string, data X,
	run func(string, X, chan HostingPlaceData[Y])) map[string]Y {
	result := make(map[string]Y)
	c := make(chan HostingPlaceData[Y], len(hps))
	var wg sync.WaitGroup
	for _, hp := range hps {
		wg.Add(1)
		go func(region string) {
			defer wg.Done()
			run(region, data, c)
		}(hp)
	}
	go func() {
		wg.Wait()
		close(c)
	}()
	for rr := range c {
		if rr.Err == nil {
			result[rr.Region] = rr.Value
		}
	}
	return result
}
