package rhelai

import (
	cr "github.com/redhat-developer/mapt/pkg/provider/api/compute-request"
	spotTypes "github.com/redhat-developer/mapt/pkg/provider/api/spot"
)

type RHELAIArgs struct {
	Prefix         string
	Accelerator    string
	Version        string
	CustomAMI      string
	Arch           string
	ComputeRequest *cr.ComputeRequestArgs
	Spot           *spotTypes.SpotArgs
	// If timeout is set a severless scheduled task will be created to self destroy the resources
	Timeout string
}
