package gcp

import (
	"fmt"
	"os"
)

const (
	ENV_PROJECT_ID = "GOOGLE_CLOUD_PROJECT"
)

type GCP struct{}

func (g *GCP) Init(backedURL string) error {
	return fmt.Errorf("not implemented yet")
}

func GetProjectID() string {
	return os.Getenv("GOOGLE_CLOUD_PROJECT")
}

func Provider() *GCP {
	return &GCP{}
}
