package output

import (
	"os"
	"path"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

func Write(stackResult auto.UpResult, destinationFolder string, results map[string]string) (err error) {
	for k, v := range results {
		if err = writeOutput(stackResult, k, destinationFolder, v); err != nil {
			return err
		}
	}
	return
}

func writeOutput(stackResult auto.UpResult, outputkey, destinationFolder, destinationFilename string) error {
	value, ok := stackResult.Outputs[outputkey].Value.(string)
	if ok {
		err := os.WriteFile(path.Join(destinationFolder, destinationFilename), []byte(value), 0600)
		if err != nil {
			return err
		}
	} else {
		logging.Debugf("error getting %s", outputkey)
	}
	return nil
}
