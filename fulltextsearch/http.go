package fulltextsearch

import (
	"log"
	"net/http"

	"github.com/meidomx/misc-service/config"

	"github.com/blevesearch/bleve"
	"github.com/gin-gonic/gin"
)

type CommonDocument struct {
	Id   string `json:"doc_id" form:"doc_id"`
	Body string `json:"body,omitempty" form:"body"`
}

func SearchDocument(container *config.Container) func(context *gin.Context) {
	return func(ctx *gin.Context) {
		v := ctx.Query("q")
		if len(v) == 0 {
			ctx.JSON(http.StatusNoContent, ResponseOk())
			return
		}

		query := bleve.NewMatchQuery(v)
		search := bleve.NewSearchRequest(query)
		searchResults, err := container.BleveIndex.Search(search)
		if err != nil {
			log.Println("Search document error:", err)
			ctx.JSON(http.StatusInternalServerError, ResponseErr("5000", "search document error"))
			return
		}

		if searchResults.Total <= 0 {
			ctx.JSON(http.StatusNoContent, ResponseOk())
			return
		} else {
			var docList []*CommonDocument
			for _, v := range searchResults.Hits {
				docList = append(docList, &CommonDocument{
					Id: v.ID,
				})
			}
			ctx.JSON(http.StatusOK, ResponseBody(docList))
			return
		}
	}
}

func InsertOrUpdateDocument(container *config.Container) func(context *gin.Context) {
	return func(ctx *gin.Context) {
		doc := new(CommonDocument)
		if ctx.ShouldBind(&doc) == nil {
			log.Println("receive document:", doc.Id)
		}
		err := container.BleveIndex.Index(ctx.Param("doc_id"), doc)
		if err != nil {
			log.Println("insert or update document error:", err)
			ctx.JSON(http.StatusInternalServerError, ResponseErr("5000", "insert or update document error"))
		} else {
			ctx.JSON(http.StatusOK, ResponseOk())
		}
	}
}

func DeleteDocument(container *config.Container) func(context *gin.Context) {
	return func(ctx *gin.Context) {
		err := container.BleveIndex.Delete(ctx.Param("doc_id"))
		if err != nil {
			log.Println("delete document error:", err)
			ctx.JSON(http.StatusInternalServerError, ResponseErr("5000", "delete document error"))
		} else {
			ctx.JSON(http.StatusOK, ResponseOk())
		}
	}
}

func ResponseOk() map[string]interface{} {
	return map[string]interface{}{
		"status": "0",
	}
}

func ResponseBody(body interface{}) map[string]interface{} {
	return map[string]interface{}{
		"status": "0",
		"data":   body,
	}
}

func ResponseErr(errorCode, errorMessage string) map[string]interface{} {
	return map[string]interface{}{
		"status":        "1000",
		"error_code":    errorCode,
		"error_message": errorMessage,
	}
}
