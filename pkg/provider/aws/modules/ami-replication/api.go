package amireplication

import (
	"fmt"
	"os"

	"github.com/adrianriobo/qenvs/pkg/provider/aws/data"
	"github.com/adrianriobo/qenvs/pkg/util/logging"
)

const (
	operationCreate  = "create"
	operationDestroy = "destroy"
)

func CreateReplicas(projectName, backedURL,
	amiID, amiName, amiSourceRegion string) (err error) {
	return manageReplicas(projectName, backedURL, amiID, amiName, amiSourceRegion, operationCreate)
}

func DestroyReplicas(projectName, backedURL string) (err error) {
	return manageReplicas(projectName, backedURL, "", "", "", operationDestroy)
}

func manageReplicas(projectName, backedURL,
	amiID, amiName, amiSourceRegion, operation string) (err error) {

	request := ReplicatedRequest{
		ProjectName:     projectName,
		AMITargetName:   amiName,
		AMISourceID:     amiID,
		AMISourceRegion: amiSourceRegion}

	regions, err := data.GetRegions()
	if err != nil {
		logging.Errorf("failed to get regions")
		os.Exit(1)
	}
	errChan := make(chan error)
	for _, region := range regions {
		// Do not replicate on source region
		if region != amiSourceRegion {
			go request.runStackAsync(backedURL, region, operation, errChan)
		}
	}
	hasErrors := false
	for _, region := range regions {
		if region != amiSourceRegion {
			if err := <-errChan; err != nil {
				logging.Errorf("%v", err)
				hasErrors = true
			}
		}
	}
	if hasErrors {
		return fmt.Errorf("there are errors on some replications. Check the logs to get information")
	}
	return nil
}
