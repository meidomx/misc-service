package smallobj

import (
	"github.com/meidomx/misc-service/config"

	"github.com/gin-gonic/gin"
)

func InitSmallObj(c *config.Config, engine *gin.Engine, container *config.Container) {
	engine.GET("/small_obj/file_content/:doc_id", GetSmallObjectContent)
	engine.PUT("/small_obj/:doc_id", InsertSmallObject)
	engine.DELETE("/small_obj/:doc_id", DeleteSmallObject)
}
