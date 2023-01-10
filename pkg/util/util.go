package util

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func If[T any](cond bool, vtrue, vfalse T) T {
	if cond {
		return vtrue
	}
	return vfalse
}

func ArrayCast[T any](source []interface{}) []T {
	var result []T
	for _, item := range source {
		result = append(result, item.(T))
	}
	return result
}

func ArrayConvert[T any, Y any](source []Y,
	convert func(item Y) T) []T {
	var result []T
	for _, item := range source {
		result = append(result, convert(item))
	}
	return result
}

func SplitString(source, delimiter string) []string {
	splitted := strings.Split(source, delimiter)
	return If(len(splitted[0]) > 0, splitted, []string{})
}

// This function split a list based on a evaluation function
// the result is a map of lists
func SplitSlice[T any, Y comparable](source []T,
	identification func(item T) Y) (m map[Y][]T) {
	m = make(map[Y][]T)
	for _, item := range source {
		k := identification(item)
		if l, found := m[k]; found {
			m[k] = append(l, item)
		} else {
			m[k] = []T{item}
		}
	}
	return
}

func Average(source []float64) float64 {
	total := 0.0
	for _, v := range source {
		total += v
	}
	return total / float64(len(source))
}

func Max(source []float64) float64 {
	total := 0.0
	for _, v := range source {
		total += v
	}
	return total / float64(len(source))
}

func Template(data any, templateName, templateContent string) (string, error) {
	tmpl, err := template.New(templateName).Parse(templateContent)
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

func WriteTempFile(content string) (string, error) {
	tmpFile, err := ioutil.TempFile("", fmt.Sprintf("%s-", filepath.Base(os.Args[0])))
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()
	_, err = tmpFile.WriteString(content)
	return tmpFile.Name(), err
}
