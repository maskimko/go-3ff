package cmd

import (
	"fmt"
	"github.com/maskimko/go-3ff/hclparser"
	"os"
)

func TFGrep(path, pattern string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open %s %w", path, err)
	}
	body, err := hclparser.GetResourceBody(pattern, f)
	if err != nil {
		return "", nil
	}
	return body, nil
}
