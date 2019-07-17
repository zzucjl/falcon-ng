package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/go-sql-driver/mysql"

	"github.com/toolkits/pkg/file"
	"github.com/toolkits/pkg/logger"
	"github.com/toolkits/pkg/runner"

	"github.com/open-falcon/falcon-ng/src/model"
	"github.com/open-falcon/falcon-ng/src/modules/portal/config"
	"github.com/open-falcon/falcon-ng/src/modules/portal/http"
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

	runner.Init()
	fmt.Println("portal start, use configuration file:", *conf)
	fmt.Println("runner.Cwd:", runner.Cwd)
	fmt.Println("runner.Hostname:", runner.Hostname)
}

func main() {
	aconf()
	pconf()

	config.InitLogger()
	model.InitMySQL("uic", "portal", "mon")
	model.InitRoot()

	http.Start()
	ending()
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
	fmt.Println("portal stopped successfully")
}

// auto detect configuration file
func aconf() {
	if *conf != "" && file.IsExist(*conf) {
		return
	}

	*conf = "etc/portal.local.yml"
	if file.IsExist(*conf) {
		return
	}

	*conf = "etc/portal.yml"
	if file.IsExist(*conf) {
		return
	}

	fmt.Println("no configuration file for portal")
	os.Exit(1)
}

// parse configuration file
func pconf() {
	if err := config.Parse(*conf); err != nil {
		fmt.Println("cannot parse configuration file:", err)
		os.Exit(1)
	}
}
