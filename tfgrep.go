package main

import (
	"fmt"
	"github.com/maskimko/go-3ff/cmd"
	"log"
	"os"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("Usage: %s <pattern> <.tf file or dir>", os.Args[0])
	}
	query := fmt.Sprintf("resource.%s", os.Args[1])
	grep, err := cmd.TFGrep(os.Args[2], query)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println(grep)
}
