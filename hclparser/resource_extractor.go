package hclparser

import (
	"fmt"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	"github.com/maskimko/go-3ff/utils"
	"os"
	"regexp"
	"strings"
)

func GetResourceBody(resourceName string, f *os.File) (string, error) {
	resourcePattern := fmt.Sprintf("resource\\.%s", resourceName)
	return QueryBody(resourcePattern, f)
}

func QueryBody(pattern string, f *os.File) (string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("failed to compile given regex %s %w", pattern, err)
	}
	cumulativeBody, err := GetCumulativeBody(f)
	if err != nil {
		return "", err
	}
	var builder strings.Builder
	if cumulativeBody != nil {
		for _, b := range cumulativeBody.GetBlocks() {
			rn := getResourceName(b)
			if re.MatchString(rn) {
				data, err := utils.GetStringFromRange(b.Range())
				if err != nil {
					return "", fmt.Errorf("failed to fetch body of resource %s %w", rn, err)
				}
				builder.WriteString(data)
				builder.WriteRune('\n')
			}
		}
	}
	if builder.Len() == 0 {
		return "", fmt.Errorf("no resources were found by pattern %s", pattern)
	}
	return builder.String(), nil
}

func getResourceName(b *hclsyntax.Block) string {
	parts := []string{b.Type}
	parts = append(parts, b.Labels...)
	return strings.Join(parts, LabelSeparator)
}
