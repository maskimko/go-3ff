package main

import (
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/maskimko/go-3ff/hclparser"
	"github.com/maskimko/go-3ff/utils"
	"log"
	"os"
	"strings"
)

//Version info
var major string
var minor string
var revision string

type optionSet struct {
	version *bool
	debug   *bool
	trfout  *bool
	nopb    *bool
	//Source file/directoryl
	s *string
	//Modified file/directoryuint8
	m *string
}

func checkEnv(opts *optionSet) {
	//if _, ok := os.LookupEnv("TFRESDIF_NOPB"); ok {
	//	nopb := true
	//	opts.nopb = &nopb
	//}
	//TODO: use viper library here
	if _, ok := os.LookupEnv("3FF_DEBUG"); ok {
		debug := true
		opts.debug = &debug
	}
}

func main() {
	opts := parseOptions()
	checkEnv(opts)
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	//var bar *pb.ProgressBar = nil
	if *opts.debug {
		log.Printf("Program was been launched with this arguments \"%s\"", strings.Join(os.Args, " "))
	}
	hclparser.Debug = *opts.debug
	hclparser.TerraformOutput = *opts.trfout
	var res *hclparser.ModifiedResources
	if *opts.debug && *opts.trfout {
		log.Println("Warning! Debug output can spoil terraform output for consumption")
	}

	orig, err := os.Open(*opts.s)
	if err != nil {
		log.Fatalf("Cannot open file %s. Error: %s", *opts.s, err)
	}
	modif, err := os.Open(*opts.m)
	res, err = hclparser.CompareFiles(orig, modif)
	if err != nil {
		log.Fatalf("Cannot open file %s. Error: %s", *opts.m, err)
	}

	if *opts.trfout {
		for _, r := range *res.List() {
			if strings.HasPrefix(r, "resource") || strings.HasPrefix(r, "module") || strings.HasPrefix(r, "data") {
				fmt.Println(strings.Replace(strings.Split(r, "/")[0], "resource.", "", 1))
			}
		}
	}

	if res.IsEmpty() {
		if *opts.debug {
			log.Println(green("OK! No changes"))
		}
	} else {
		if *opts.debug {
			log.Println(red("There are changes"))
		}
	}
	if *opts.debug {
		log.Printf("File cache hits: %d", utils.CacheHits)
	}
}

func parseOptions() *optionSet {
	o := optionSet{}
	o.debug = flag.Bool("d", false, "Enable debug output to StdErr")
	o.version = flag.Bool("version", false, "Show version info")
	o.trfout = flag.Bool("t", false, "Output modified resources only (For terraform command)")
	//o.nopb = flag.Bool("nopb", false, "Disable displaying of progress bar")
	flag.Parse()
	if *o.version {
		fmt.Printf("Version: v%s.%s.%s\n", major, minor, revision)
		os.Exit(0)
	}

	switch len(flag.Args()) {
	case 0:
		log.Fatalf("%s You have to specify files or directories for comparison", color.RedString("Error:"))
	case 1:
		if *o.debug {
			log.Println("If you specify only one argument as source directory outside git.go context, the current directory will be treated as a modified file directory")
		}
		sd := flag.Arg(0)
		o.s = &sd
		wd, err := os.Getwd()
		if err != nil {
			log.Fatalf("Cannot get cuurrent working directory: %s", err)
		}
		o.m = &wd
	case 2:
		sd := flag.Arg(0)
		o.s = &sd
		md := flag.Arg(1)
		o.m = &md
	default:
		log.Fatalln("Outside git.go context you must specify two arguments <source> <modified> file/directory to perform diff calculation")
	}
	return &o
}
