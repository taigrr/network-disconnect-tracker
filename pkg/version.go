package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

const (
	Package = "network-disconnect-tracker"
)

var (
	GitCommit string
	Tag       string
	BuildTime string
	Authors   string
	BuildNo   string
	version   = flag.Bool("version", false, "Get detailed version string")
	process   = flag.String("p", "unk", "Specify the process logging for filename")
	debug     = flag.Bool("debug", false, "Enable debug print")
)

func init() {
	flag.Parse()

	Authors = strings.ReplaceAll(Authors, "SpAcE", " ")
	Tag = strings.ReplaceAll(Tag, ";", "; ")

	if GitCommit == "" || Tag == "" || BuildTime == "" {
		log.Fatalf("Binary built improperly. Version variables not set!")
	}
	if *version {
		fmt.Printf("%s Version information:\n|| Authors: %s\n|| Commit: %s\n|| Tag: %s\n|| Build No: %s\n|| Build Date: %s\n", Package, Authors, GitCommit, Tag, BuildNo, BuildTime)
		os.Exit(0)
	}
}
