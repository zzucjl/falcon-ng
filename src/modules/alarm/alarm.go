package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/go-sql-driver/mysql"
	"github.com/toolkits/pkg/file"
	"github.com/toolkits/pkg/logger"
	"github.com/toolkits/pkg/runner"

	"github.com/open-falcon/falcon-ng/src/model"
	"github.com/open-falcon/falcon-ng/src/modules/alarm/cache"
	"github.com/open-falcon/falcon-ng/src/modules/alarm/config"
	"github.com/open-falcon/falcon-ng/src/modules/alarm/cron"
	"github.com/open-falcon/falcon-ng/src/modules/alarm/http"
	"github.com/open-falcon/falcon-ng/src/modules/alarm/redi"
)

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
		fmt.Println("version:", config.Version)
		os.Exit(0)
	}

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	runner.Init()
	fmt.Println("alarm start, use configuration file:", *conf)
	fmt.Println("runner.Cwd:", runner.Cwd)
	fmt.Println("runner.Hostname:", runner.Hostname)
}

func main() {
	aconf()
	pconf()
	cache.Init()

	config.InitLogger()
	model.InitMySQL("uic", "portal", "mon")

	if err := cron.SyncMaskconf(); err != nil {
		log.Fatalf("sync mask failed, err: %v", err)
	}

	if err := cron.SyncStra(); err != nil {
		log.Fatalf("sync stra failed, err: %v", err)
	}

	redi.InitRedis()
	go cron.SyncMaskconfLoop()
	go cron.SyncStraLoop()
	go cron.ReadHighEvent()
	go cron.ReadLowEvent()
	go cron.CallbackConsumer()
	go cron.MergeEvent()
	go cron.CleanEventLoop()

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
	redi.CloseRedis()
	fmt.Println("alarm stopped successfully")
}

// auto detect configuration file
func aconf() {
	if *conf != "" && file.IsExist(*conf) {
		return
	}

	*conf = "etc/alarm.local.yml"
	if file.IsExist(*conf) {
		return
	}

	*conf = "etc/alarm.yml"
	if file.IsExist(*conf) {
		return
	}

	fmt.Println("no configuration file for alarm")
	os.Exit(1)
}

// parse configuration file
func pconf() {
	if err := config.ParseCfg(*conf); err != nil {
		fmt.Println("cannot parse configuration file:", err)
		os.Exit(1)
	}
}
