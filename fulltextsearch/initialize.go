package fulltextsearch

import (
	"fmt"
	"log"
	"os"

	"github.com/meidomx/misc-service/config"

	"github.com/blevesearch/bleve"
	"github.com/gin-gonic/gin"
	"github.com/spf13/afero"
)

func InitService(c *config.Config, engine *gin.Engine, container *config.Container) {
	filename := "single_index.bleve"
	fullPath := fmt.Sprint(c.FullTextSearch.IndexFolder, string(os.PathSeparator), filename)

	exists := false
	fs := afero.NewOsFs()
	if ok, err := afero.Exists(fs, fullPath); err != nil {
		log.Println("check full text search index folder exist error:", err)
		panic(err)
	} else if !ok {
		if err := fs.MkdirAll(c.FullTextSearch.IndexFolder, 0755); err != nil {
			log.Println("create full text search index folder error:", err)
			panic(err)
		}
	} else {
		exists = true
	}

	if exists {
		idx, err := bleve.Open(fullPath)
		if err != nil {
			log.Println("open full text search index error:", err)
			panic(err)
		}
		container.BleveIndex = idx
	} else {
		mapping := bleve.NewIndexMapping()
		idx, err := bleve.New(fullPath, mapping)
		if err != nil {
			log.Println("new full text search index error:", err)
			panic(err)
		}
		container.BleveIndex = idx
	}

	fmt.Println("start bleve index at:", fullPath)

	engine.GET("/full_text/documents", SearchDocument(container))
	engine.PUT("/full_text/document/:doc_id", InsertOrUpdateDocument(container))
	engine.DELETE("/full_text/document/:doc_id", DeleteDocument(container))
}
