package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/open-falcon/falcon-ng/src/modules/logcollector/config"
	"github.com/open-falcon/falcon-ng/src/modules/logcollector/http"
	"github.com/open-falcon/falcon-ng/src/modules/logcollector/worker"

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
	config.Hostname, err = config.GetHostname()
	if err != nil {
		logger.Fatal("cannot get hostname:", err)
	} else {
		logger.Info("hostname:", config.Hostname)
	}

	if config.Hostname == "127.0.0.1" {
		log.Fatalln("hostname: 127.0.0.1, cannot work")
	}

	go worker.UpdateConfigsLoop() // step2, step1和step2有顺序依赖
	go worker.PusherStart()
	go worker.Zeroize()

	http.Start()
	ending()
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

// parse configuration file
func pconf() {
	if err := config.Parse(*conf); err != nil {
		fmt.Println("cannot parse configuration file:", err)
		os.Exit(1)
	}
}

func start() {
	runner.Init()
	fmt.Println("logcollector start, use configuration file:", *conf)
	fmt.Println("runner.Cwd:", runner.Cwd)
	fmt.Println("runner.Hostname:", runner.Hostname)
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
