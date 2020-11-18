package hclparser

import (
	"errors"
	"fmt"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	"github.com/hashicorp/hcl2/hclparse"
	"github.com/maskimko/go-3ff/utils"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

var Debug bool = false
var Logger *log.Logger

type ResourceType string

const (
	ResourceKey    ResourceType = "resource"
	DataKey        ResourceType = "data"
	ModuleKey      ResourceType = "module"
	OutputKey      ResourceType = "output"
	LabelSeparator              = "."
	TerraformKey   ResourceType = "terraform"
	VariableKey    ResourceType = "variable"
	ProviderKey    ResourceType = "provider"
)

/**
Compare function performs comparison of 2 files, which it receives as arguments, and returns true if there are no diff
o stands for original, m stands for modified
It can compare terrafrom files in HCL2 format only
If arguments are names of directories, it will try to perform File-by-File comparison
*/
func Compare(o, m string) (*ModifiedResources, error) {
	oFile, err := os.Open(o)
	if err != nil {
		log.Fatalf("Cannot open File %s", o)
	}
	defer oFile.Close()
	mFile, err := os.Open(m)
	if err != nil {
		log.Fatalf("Cannot open File %s", m)
	}
	defer mFile.Close()
	return CompareFiles(oFile, mFile)
}

//CompareFiles function can read data from given files and perform a comparison of Terraform data
func CompareFiles(o, m *os.File) (*ModifiedResources, error) {
	ofi, err := o.Stat()
	if err != nil {
		if Debug {
			Logger.Printf("Cannot get File info of %s. Error message: %s", o.Name(), err)
		}
		return nil, err
	}
	mfi, err := m.Stat()
	if err != nil {
		if Debug {
			Logger.Printf("Cannot get File info of %s. Error message: %s", m.Name(), err)
		}
		return nil, err
	}
	if ofi.IsDir() != mfi.IsDir() {
		if Debug {
			Logger.Printf("Both files you specified, should be directories, or both should be files\n%s %s", o.Name(), m.Name())
		}
		return nil, errors.New("error: different file types: both files you specified, should be directories, or both should be files")
	}
	mr := NewModifiedResources()
	origFiles, err := getFilesSlice(o)
	if err != nil {
		if Debug {
			Logger.Printf("Cannot build files list of the directory %s. Error: %s", ofi.Name(), err)
		}
		return nil, err
	}
	for _, ofc := range origFiles {
		defer ofc.File.Close()
	}
	modifFiles, err := getFilesSlice(m)
	if err != nil {
		if Debug {
			Logger.Printf("Cannot build files list of the directory %s. Error: %s", mfi.Name(), err)
		}
		return nil, err
	}
	for _, mfc := range origFiles {
		defer mfc.File.Close()
	}
	ohf, err := getHclFiles(origFiles)
	if err != nil {
		if Debug {
			Logger.Printf("Cannot parse original files. Error %s", err)
		}
		return nil, err
	}
	mhf, err := getHclFiles(modifFiles)
	if err != nil {
		if Debug {
			Logger.Printf("Cannot parse modified files %s", err)
		}
		return nil, err
	}
	origCumulativeBody := unpack(ohf)
	modifCumulativeBody := unpack(mhf)

	mr.computeBodyDiff(origCumulativeBody, modifCumulativeBody, nil)
	return mr, err
}

//getFilesSlice function returns a sorted by file name list of files, which was generated by walking through the given directory
func getFilesSlice(root *os.File) (SortableFiles, error) {
	if root == nil {
		return nil, nil
	}
	fileInfo, err := root.Stat()
	if err != nil {
		if Debug {
			Logger.Printf("Cannot stat File %s", root.Name())
		}
		return nil, err
	}
	if !fileInfo.IsDir() {
		return SortableFiles{SortableFile{File: root}}, nil
	} else {
		var fl SortableFiles
		err := filepath.Walk(root.Name(), func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				if strings.HasPrefix(info.Name(), ".") {
					return filepath.SkipDir
				}
			}
			if strings.HasSuffix(path, ".tf") {
				f, err := os.Open(path)
				if err != nil {
					if Debug {
						Logger.Printf("Cannot open File %s", path)
					}
					return err
				}
				fl = append(fl, SortableFile{File: f})
			}
			return nil
		})
		if err != nil {
			if Debug {
				Logger.Printf("Cannot walk the directory "+
					"%s tree. Error: %s", root.Name(), err)
			}
			return nil, err
		}
		sort.Sort(fl)
		return fl, nil
	}
}

//getHclFiles function return an array of HCL parsed data from given files
func getHclFiles(o SortableFiles) ([]*hcl.File, error) {
	var allFiles []*hcl.File = make([]*hcl.File, len(o))
	parser := hclparse.NewParser()
	for i, sf := range o {
		bytes, err := ioutil.ReadAll(sf.File)
		if err != nil {
			log.Fatalf("Cannot read File %s", sf.File.Name())
			return nil, err
		}
		hclFile, diag := parser.ParseHCL(bytes, sf.File.Name())
		if diag != nil && diag.HasErrors() {
			for _, err := range diag.Errs() {
				Logger.Printf("Cannot parse File %s. Error: %s", sf.File.Name(), err)
				return nil, err
			}
		}
		//By using explicit index I maintain the files order
		allFiles[i] = hclFile
	}
	//NOTE: Perhaps it worth to make diff of the files and output it somehow.
	// Though it is not directly related to the terraform resources
	return allFiles, nil
}

//unpack function returns body of the HCL file
func unpack(hfls []*hcl.File) *Body {
	var atr hclsyntax.Attributes = make(map[string]*hclsyntax.Attribute)
	var hclb hclsyntax.Body = hclsyntax.Body{Attributes: atr, Blocks: make([]*hclsyntax.Block, 0)}
	for _, f := range hfls {
		var hb *hclsyntax.Body = f.Body.(*hclsyntax.Body)
		for k, v := range hb.Attributes {
			if hclb.Attributes[k] != nil {
				if Debug {
					//Check for duplicates
					Logger.Printf("Cummulative attributes map already contains the value for the key %s", k)
				}
			}
			hclb.Attributes[k] = v
		}
		for _, b := range hb.Blocks {
			hclb.Blocks = append(hclb.Blocks, b)
		}
	}
	b := Body(hclb)
	return &b
}

//GetTfResourcesByPath function returns all parsed terraform resource names, which can be used as targets in terraform
func GetTfResourcesByPath(path string) ([]string, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if Debug {
			log.Printf("Cannot find file %s Error: %s", path, err)
		}
		return nil, err
	} else {
		file, err := os.Open(path)
		if err != nil {
			if Debug {
				log.Printf("Cannot open file %s Error: %s", path, err)
			}
			return nil, err
		}
		return GetTfResources(file)
	}
}

//GetTfResources function returns all parsed terraform resource names, which can be used as targets in terraform
func GetTfResourcesCount(s *os.File) (map[string]int, error) {
	sfi, err := s.Stat()
	if err != nil {
		if Debug {
			Logger.Printf("Cannot get File info of %s. Error message: %s", s.Name(), err)
		}
		return nil, err
	}
	files, err := getFilesSlice(s)
	if err != nil {
		if Debug {
			Logger.Printf("Cannot build files list of the directory %s. Error: %s", sfi.Name(), err)
		}
		return nil, err
	}
	for _, ofc := range files {
		defer ofc.File.Close()
	}

	hf, err := getHclFiles(files)
	if err != nil {
		if Debug {
			Logger.Printf("Cannot parse original files. Error %s", err)
		}
		return nil, err
	}
	cumulativeBody := unpack(hf)
	var resources = make(map[string]int)
	if cumulativeBody != nil {
		for n, b := range cumulativeBody.GetBlocks() {
			if Debug {
				log.Printf("===processing block #%d===", n)
			}
			addResource(b, resources)
		}
	}
	return resources, err
}

func getCount(b *hclsyntax.Block) int {
	attributes := b.Body.Attributes
	//attrMap := attributes.(map[string]*hclsyntax.Attribute)
	for k, v := range attributes {
		if k == "count" {
			val, d := v.Expr.Value(nil)
			if Debug && d != nil && d.HasErrors() {
				log.Printf(d.Error())
			}
			if val.IsWhollyKnown() && val.Type().IsPrimitiveType() {
				switch val.Type().GoString() {
				case "cty.String":
					i, err := strconv.Atoi(val.AsString())
					if err != nil {
						if Debug {
							log.Printf("failed to convert %s to integer", val.AsString())
						}
						return -1
					}
					return i
				case "cty.Number":
					bf := val.AsBigFloat()
					if bf.IsInt() {
						bi, _ := bf.Int(nil)
						i := int(bi.Int64())
						return i
					} else {
						return -1
					}
				default:
					if Debug {
						log.Printf("got invlid count value at %s line %d col %d", v.Expr.Range().Filename,
							v.Expr.Range().Start.Line, v.Expr.Range().Start.Column)
					}
					return -1
				}

			}
		}
	}
	return 1
}

func addResource(b *hclsyntax.Block, resources map[string]int) {
	switch b.Type {
	case string(ResourceKey):
		//resources = append(resources, strings.Join(b.Labels, LabelSeparator))

		resources[strings.Join(b.Labels, LabelSeparator)] = getCount(b)
	case string(ModuleKey):
		//resources = append(resources, string(ModuleKey)+LabelSeparator+strings.Join(b.Labels, LabelSeparator))
		resources[string(ModuleKey)+LabelSeparator+strings.Join(b.Labels, LabelSeparator)] = 1
	case string(DataKey):
		//resources = append(resources, string(DataKey)+LabelSeparator+strings.Join(b.Labels, LabelSeparator))
		resources[string(DataKey)+LabelSeparator+strings.Join(b.Labels, LabelSeparator)] = 1
	case string(OutputKey):
		//resources = append(resources, string(OutputKey)+LabelSeparator+strings.Join(b.Labels, LabelSeparator))
		resources[string(DataKey)+LabelSeparator+strings.Join(b.Labels, LabelSeparator)] = 1
	case string(TerraformKey):
		//Skip terraform definition
	case string(VariableKey):
		//resources = append(resources, string(VariableKey)+LabelSeparator+strings.Join(b.Labels, LabelSeparator))
		resources[string(VariableKey)+LabelSeparator+strings.Join(b.Labels, LabelSeparator)] = 1
	case string(ProviderKey):
		//Skip the provider definition
	default:
		if Debug {
			log.Printf("Unsupported type of resource %s", b.Type)
		}
	}
}

func GetTfResources(s *os.File) ([]string, error) {
	var resources []string
	count, err := GetTfResourcesCount(s)
	if err != nil {
		return nil, err
	}
	for k, _ := range count {
		resources = append(resources, k)
	}
	return resources, nil
}

//computeBodyDiff function computes diff of two HCL bodies and prints computed data
func (mr *ModifiedResources) computeBodyDiff(ob, mb *Body, path []string) bool {
	oAttrs := NewAttributes(ob.Attributes)
	mAttrs := NewAttributes(mb.Attributes)
	oBlocks := ob.GetBlocks()
	mBlocks := mb.GetBlocks()
	printParams := GetDefaultPrintParams()
	atdf := mr.analyzeAttributesDiff(oAttrs, mAttrs, path)
	if atdf.HasChanges() {
		PrintModified(strings.Join(path, "/"), printParams)
		printParams.Shift()
		PrintAttributeContext(atdf, printParams)
		printParams.Unshift()
		return false
	}
	return mr.analyzeBlocksDiff(oBlocks, mBlocks, path, GetDefaultPrintParams())
}

//computeBlockDiff function returns true if blocks are equal
func (mr *ModifiedResources) computeBlockDiff(o, m *Block, path []string) bool {
	//p := append(path, fmt.Sprintf("%s.%s", o.Type, strings.Join(o.Labels, ".")))
	var pChunk string
	if o.Labels != nil && len(o.Labels) > 0 {
		pChunk = fmt.Sprintf("%s.%s", o.Type, strings.Join(o.Labels, "."))
	} else {
		pChunk = o.Type
	}
	p := append(path, pChunk)
	if o.Type != m.Type {
		if Debug {
			Logger.Printf("Block types differ. Path: %s\n"+
				"                    Original: %s (in File %s at line: %d, column: %d)\n"+
				"                    Modified: %s (in File %s at line: %d, column: %d)", strings.Join(path, "."),
				o.Type, o.TypeRange.Filename, o.TypeRange.Start.Line, o.TypeRange.Start.Column,
				m.Type, m.TypeRange.Filename, m.TypeRange.Start.Line, m.TypeRange.Start.Column)
			logString, err := utils.GetChangeLogString(o.TypeRange, m.TypeRange)
			if err != nil {
				Logger.Print("Cannot compose type diff")
			} else {
				Logger.Println(logString)
			}

		}
		mr.Add(strings.Join(p, "/"))
		return false
	}

	if !mr.computeLabelsDiff(o, m, p) {
		return false
	}

	return mr.computeBodyDiff(o.GetBody(), m.GetBody(), p)

}

//computeLabelsDiff function returns true if labels are equal
func (mr *ModifiedResources) computeLabelsDiff(o, m *Block, path []string) bool {
	if len(o.Labels) != len(m.Labels) {
		if Debug {

			//Basically this case should never happen
			Logger.Println("WARNING!!! This should never happen!")
			Logger.Printf("Lables quantity differ. Path: %s\n"+
				"                    Original: %d (in File %s)\n"+
				"                    Modified: %d (in File %s)", strings.Join(path, "/"),
				len(o.Labels), o.Range().Filename,
				len(m.Labels), m.Range().Filename)
			logString, err := utils.GetChangeLogString(o.Range(), m.Range())
			if err != nil {
				Logger.Print("Cannot compose labels quantity diff")
			} else {
				Logger.Println(logString)
			}
		}

		return false
	}
	for i, v := range o.Labels {
		if v != m.Labels[i] {
			if Debug {
				Logger.Printf("Lables  differ. Path: %s\n"+
					"                    Original: %s (in File %s at line: %d, column: %d)\n"+
					"                    Modified: %s (in File %s at line: %d, column: %d)", strings.Join(path, "/"),
					o.Type, o.LabelRanges[i].Filename, o.LabelRanges[i].Start.Line, o.LabelRanges[i].Start.Column,
					m.Type, m.LabelRanges[i].Filename, m.LabelRanges[i].Start.Line, m.LabelRanges[i].Start.Column)
				logString, err := utils.GetChangeLogString(o.LabelRanges[i], m.LabelRanges[i])
				if err != nil {
					Logger.Print("Cannot compose label diff")
				} else {
					Logger.Println(logString)
				}
			}
			mr.Add(strings.Join(path, "/"))
			return false
		}
	}
	return true
}
