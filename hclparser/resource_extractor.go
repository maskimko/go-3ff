package hclparser

import (
	"fmt"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	"github.com/maskimko/go-3ff/utils"
	"os"
	"regexp"
	"strings"
	"sync"
)

//Deprecated use Extractor interface instead
func GetResourceBody(resourceName string, f *os.File) (string, error) {
	resourcePattern := fmt.Sprintf("resource\\.%s", resourceName)
	return QueryBody(resourcePattern, f)
}

//Deprecated use Extractor interface instead
func QueryBody(pattern string, f *os.File) (string, error) {
	var extractor Extractor = NewDefaultResourceExtractor()
	return extractor.QueryBody(pattern, f)

}

func getResourceName(b *hclsyntax.Block) string {
	parts := []string{b.Type}
	parts = append(parts, b.Labels...)
	return strings.Join(parts, LabelSeparator)
}

/*
Use can use NewDefaultResourceExtractor function to create a default instance of the Extractor
*/
type Extractor interface {
	QueryBody(pattern string, f *os.File) (string, error)
	ResetCache()
}

type ResourceExtractorImpl struct {
	lock  sync.Mutex
	cache map[string]*Body
}

func NewDefaultResourceExtractor() *ResourceExtractorImpl {
	return &ResourceExtractorImpl{
		cache: make(map[string]*Body),
	}
}
func (r *ResourceExtractorImpl) QueryBody(pattern string, f *os.File) (string, error) {
	if r.cache == nil {
		r.ResetCache()
	}
	if f == nil {
		return "", fmt.Errorf("nil value file reference provided")
	}
	fname := f.Name()
	var err error
	if _, ok := r.cache[fname]; !ok {
		r.lock.Lock()
		if _, doubleOK := r.cache[fname]; doubleOK {
			r.cache[fname], err = GetCumulativeBody(f)
			if err != nil {
				return "", err
			}
		}
		r.lock.Unlock()
	}
	return r.performBodyQuery(pattern, fname)
}

func (r *ResourceExtractorImpl) performBodyQuery(pattern, fname string) (string, error) {
	var builder strings.Builder
	queryRE, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("failed to compile given regex %s %w", pattern, err)
	}
	if cumulativeBody, ok := r.cache[fname]; ok && cumulativeBody != nil {
		for _, b := range cumulativeBody.GetBlocks() {
			rn := getResourceName(b)
			if queryRE.MatchString(rn) {
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

func (r *ResourceExtractorImpl) ResetCache() {
	r.lock.Lock()
	r.cache = make(map[string]*Body)
	r.lock.Unlock()
}
