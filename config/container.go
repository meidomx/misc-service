package config

import (
	"github.com/blevesearch/bleve"
	"github.com/jimlambrt/gldap"
)

type Container struct {
	GldapServer *gldap.Server
	BleveIndex  bleve.Index
}

func (this *Container) Stop() {
	if this.GldapServer != nil {
		this.GldapServer.Stop()
	}
	if this.BleveIndex != nil {
		this.BleveIndex.Close()
	}
}
