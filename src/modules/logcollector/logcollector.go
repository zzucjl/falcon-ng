package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/toolkits/pkg/file"
	"github.com/toolkits/pkg/runner"
)

const version = 1

var (
	vers *bool
	help *bool
	conf *string
)

func init() {
	vers = flag.Bool("v", false, "display the version.")
	help = flag.Bool("h", false, "print this help.")
	conf = flag.String("f", "", "specify configuration file.")
	flag.Parse()

	if *vers {
		fmt.Println("version:", version)
		os.Exit(0)
	}

	if *help {
		flag.Usage()
		os.Exit(0)
	}
}

// auto detect configuration file
func aconf() {
	if *conf != "" && file.IsExist(*conf) {
		return
	}

	*conf = "etc/logcollector.local.yml"
	if file.IsExist(*conf) {
		return
	}

	*conf = "etc/logcollector.yml"
	if file.IsExist(*conf) {
		return
	}

	fmt.Println("no configuration file for logcollector")
	os.Exit(1)
}

func main() {
	aconf()
	start()
}

func start() {
	runner.Init()
	fmt.Println("logcollector start, use configuration file:", *conf)
	fmt.Println("runner.Cwd:", runner.Cwd)
	fmt.Println("runner.Hostname:", runner.Hostname)
}
