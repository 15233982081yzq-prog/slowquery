package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	srvHttp "smart-slowquery/pkg/http/router"
	"time"
)

const (
	OsEnv = "ENV"

	ConfigFileTemplate = "/etc/conf/config.%s.toml"
)

var (
	config = flag.String("config", "", "slow query open_api config file")
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

	sv := srvHttp.NewHttpOpenApiServer(*config)
	if err := sv.Start(); err != nil {
		fmt.Printf("openapi start error:%s \n", err.Error())
	}
}
