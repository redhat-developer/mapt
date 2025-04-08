package file

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/redhat-developer/mapt/pkg/util/logging"
)

func WriteTempFile(content string) (string, error) {
	tmpFile, err := os.CreateTemp("", fmt.Sprintf("%s-", filepath.Base(os.Args[0])))
	if err != nil {
		return "", err
	}
	defer func() {
		if err := tmpFile.Close(); err != nil {
			logging.Error(err)
		}
	}()
	_, err = tmpFile.WriteString(content)
	return tmpFile.Name(), err
}

func Template(data any, templateContent string) (string, error) {
	tmpl, err := template.New("tpl").Parse(templateContent)
	if err != nil {
		return "", err
	}
	buffer := new(bytes.Buffer)
	err = tmpl.Execute(buffer, data)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}
