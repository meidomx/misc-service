package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"

	"github.com/meidomx/misc-service/config"
	"github.com/meidomx/misc-service/fulltextsearch"
	"github.com/meidomx/misc-service/id"
	"github.com/meidomx/misc-service/ldap"
	"github.com/meidomx/misc-service/pgbackend"
	"github.com/meidomx/misc-service/smallobj"

	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
)

func main() {
	c := new(config.Config)
	loadConfig(c)

	idGen := id.NewIdGen(1, 1)
	engine := gin.New()
	{
		engine.Use(gin.Logger())
		engine.Use(gin.Recovery())
	}
	container := new(config.Container)

	if err := pgbackend.InitDb(c.Storage.ConnString); err != nil {
		panic(err)
	}

	ldap.StartService(idGen, c, container)
	fulltextsearch.InitService(c, engine, container)
	smallobj.InitSmallObj(c, engine, container)

	go func() {
		// stop server gracefully when ctrl-c, sigint or sigterm occurs
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		select {
		case <-ctx.Done():
			log.Printf("\nstopping services")
			container.Stop()
		}
	}()

	if err := engine.Run(c.Http.Address); err != nil {
		panic(errors.New(fmt.Sprint("run gin failed:", err.Error())))
	}
}

func loadConfig(c *config.Config) {
	f, err := os.Open("config.toml")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	if err := toml.Unmarshal(data, c); err != nil {
		panic(err)
	}
}
