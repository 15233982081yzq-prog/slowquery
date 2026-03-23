package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	srvHttp "smart-slowquery/pkg/http/router"
)

const (
	OsEnv = "ENV"

	ConfigFileTemplate = "/etc/conf/config.%s.toml"
)

var (
	config = flag.String("config", "", "slow query alert config file")
)

func main() {
	// refresh rand seed
	// use Unix() instead of UnixNano() to leave some more space to the rand to play with.
	flag.Parse()

	env := os.Getenv(OsEnv)
	if len(env) != 0 {
		*config = fmt.Sprintf(ConfigFileTemplate, env)
	}

	// refresh rand seed
	// use Unix() instead of UnixNano() to leave some more space to the rand to play with.
	rand.Seed(time.Now().Unix())

	sv := srvHttp.NewHttpAlertServer(*config)
	if err := sv.Start(); err != nil {
		fmt.Printf("slow query alert server start error:%s \n", err.Error())
	}
}
