package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/open-falcon/falcon-ng/src/model"
	"github.com/open-falcon/falcon-ng/src/modules/syscollector/config"
	"github.com/open-falcon/falcon-ng/src/modules/syscollector/cron"
	"github.com/open-falcon/falcon-ng/src/modules/syscollector/funcs"
	"github.com/open-falcon/falcon-ng/src/modules/syscollector/http"
	"github.com/open-falcon/falcon-ng/src/modules/syscollector/plugins"
	"github.com/open-falcon/falcon-ng/src/modules/syscollector/ports"
	"github.com/open-falcon/falcon-ng/src/modules/syscollector/procs"

	"github.com/toolkits/pkg/file"
	"github.com/toolkits/pkg/logger"
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

func main() {
	aconf()
	pconf()
	start()

	var err error
	config.InitLogger()
	config.Hostname, err = config.GetEndpoint()
	if err != nil {
		logger.Fatal("cannot get endpoint:", err)
	} else {
		logger.Info("endpoint ->", config.Hostname)
	}

	if config.Hostname == "127.0.0.1" {
		log.Fatalln("hostname: 127.0.0.1, cannot work")
	}

	config.Collect = *model.NewCollect()
	funcs.BuildMappers()
	funcs.Collect()
	cron.GetCollects()

	plugins.Detect()
	procs.Detect()
	ports.Detect()

	http.Start()
	ending()
}

// auto detect configuration file
func aconf() {
	if *conf != "" && file.IsExist(*conf) {
		return
	}

	*conf = "etc/syscollector.local.yml"
	if file.IsExist(*conf) {
		return
	}

	*conf = "etc/syscollector.yml"
	if file.IsExist(*conf) {
		return
	}

	fmt.Println("no configuration file for syscollector")
	os.Exit(1)
}

// parse configuration file
func pconf() {
	if err := config.Parse(*conf); err != nil {
		fmt.Println("cannot parse configuration file:", err)
		os.Exit(1)
	}
}

func ending() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	select {
	case <-c:
		fmt.Printf("stop signal caught, stopping... pid=%d\n", os.Getpid())
	}

	logger.Close()
	http.Shutdown()
	fmt.Println("sender stopped successfully")
}

func start() {
	runner.Init()
	fmt.Println("syscollector start, use configuration file:", *conf)
	fmt.Println("runner.Cwd:", runner.Cwd)
	fmt.Println("runner.Hostname:", runner.Hostname)
}
