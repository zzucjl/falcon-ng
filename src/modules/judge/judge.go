package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/open-falcon/falcon-ng/src/modules/judge/logger"
	"github.com/open-falcon/falcon-ng/src/modules/judge/worker"

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

	*conf = "etc/judge.local.yml"
	if file.IsExist(*conf) {
		return
	}

	*conf = "etc/judge.yml"
	if file.IsExist(*conf) {
		return
	}

	fmt.Println("no configuration file for judge")
	os.Exit(1)
}

func main() {
	aconf()
	start()

	worker.Start(*conf)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	select {
	case <-c:
		logger.Info(0, "stop signal caught, try to stop judge server")
		worker.Stop()
	}
	logger.Info(0, "judge server stopped succefully")
	logger.Close()
}

func start() {
	runner.Init()
	fmt.Println("judge start, use configuration file:", *conf)
	fmt.Println("runner.Cwd:", runner.Cwd)
	fmt.Println("runner.Hostname:", runner.Hostname)
}
