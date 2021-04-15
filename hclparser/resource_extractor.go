package hclparser

import (
	"fmt"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	"github.com/maskimko/go-3ff/utils"
	"os"
	"strings"
)

func GetResourceBody(resourceName string, f *os.File) (string, error) {
	cumulativeBody, err := GetCumulativeBody(f)
	if err != nil {
		return "", err
	}
	if cumulativeBody != nil {
		for _, b := range cumulativeBody.GetBlocks() {
			rn := getResourceName(b)
			if resourceName == rn {
				data, err := utils.GetStringFromRange(b.Range())
				if err != nil {
					return "", fmt.Errorf("failed to fetch body of resource %s %w", rn, err)
				}
				return data, nil
			}
		}
	}
	return "", fmt.Errorf("no resources were found by name %s", resourceName)
}

func getResourceName(b *hclsyntax.Block) string {
	parts := []string{b.Type}
	parts = append(parts, b.Labels...)
	return strings.Join(parts, LabelSeparator)
}
