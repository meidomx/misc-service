package main

import (
	"io/ioutil"
	"os"

	"github.com/meidomx/misc-service/id"
	"github.com/meidomx/misc-service/ldap"
	"github.com/meidomx/misc-service/pgbackend"

	"github.com/BurntSushi/toml"
)

type tConfig struct {
	Storage struct {
		ConnString string `toml:"conn_string"`
	} `toml:"storage"`
}

func main() {
	c := new(tConfig)
	loadConfig(c)

	idgen := id.NewIdGen(1, 1)

	if err := pgbackend.InitDb(c.Storage.ConnString); err != nil {
		panic(err)
	}

	ldap.StartService(idgen, "dc=auth,dc=moetang,dc=com")
}

func loadConfig(c *tConfig) {
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
