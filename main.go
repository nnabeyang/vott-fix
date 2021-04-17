package main

import (
	"flag"
	"fmt"
	"log"
)

var version = "0.0.0"

func main() {
	var showVersion bool
	var keyPath string
	var targetPath string
	flag.BoolVar(&showVersion, "v", false, "show version")
	flag.BoolVar(&showVersion, "version", false, "show version")
	flag.StringVar(&targetPath, "target", "", "target path")
	flag.StringVar(&keyPath, "key-file", "", "security token file")
	flag.Parse()
	if showVersion {
		fmt.Println("version:", version)
		return
	}
	cxt, err := createVottContext(keyPath, targetPath)
	if err != nil {
		log.Fatal(err)
	}
	if err := cxt.fix(targetPath); err != nil {
		log.Fatal(err)
	}
	if err := cxt.save(); err != nil {
		log.Fatal(err)
	}
}
