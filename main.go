package main

import (
	"io/ioutil"
	"os"

	"github.com/meidomx/misc-service/id"
	"github.com/meidomx/misc-service/ldap"
	"github.com/meidomx/misc-service/pgbackend"

	"github.com/BurntSushi/toml"
)

func main() {
	c := new(config.Config)
	loadConfig(c)

	idgen := id.NewIdGen(1, 1)

	if err := pgbackend.InitDb(c.Storage.ConnString); err != nil {
		panic(err)
	}

	ldap.StartService(idgen, c)
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
