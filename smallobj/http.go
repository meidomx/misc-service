package smallobj

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"

	"github.com/meidomx/misc-service/pgbackend"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
)

const SmallObjectServiceName = "small_object"

func GetSmallObjectContent(ctx *gin.Context) {
	docId := ctx.Param("doc_id")

	var docData []byte
	if _, err := pgbackend.RunQuery(SmallObjectServiceName, &docData, func(conn *pgxpool.Conn, result *[]byte) error {
		rows, err := conn.Query(context.Background(),
			"select file_content from misc_small_object where obj_id = $1 limit 1;",
			docId,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		if rows.Next() {
			if err := rows.Scan(result); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		log.Println("get uploaded file error:", err)
		ctx.JSON(http.StatusInternalServerError, ResponseErr("5000", "internal error"))
		return
	}

	if len(docData) <= 0 {
		ctx.JSON(http.StatusNoContent, ResponseOk())
		return
	}

	ctx.Status(http.StatusOK)
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Content-Transfer-Encoding", "binary")
	ctx.Header("Content-Length", fmt.Sprintf("%d", len(docData)))
	if _, err := ctx.Writer.Write(docData); err != nil {
		log.Println("write file data error:", err.Error())
	}
}

func InsertSmallObject(ctx *gin.Context) {
	docId := ctx.Param("doc_id")
	file, err := ctx.FormFile("file")
	if err != nil {
		log.Println("get form file error:", err)
		ctx.JSON(http.StatusInternalServerError, ResponseErr("5000", "internal error"))
		return
	}
	data, err := GetUploadedFile(file)
	if err != nil {
		log.Println("get uploaded file error:", err)
		ctx.JSON(http.StatusInternalServerError, ResponseErr("5000", "internal error"))
		return
	}

	if _, err := pgbackend.RunQuery(SmallObjectServiceName, nil, func(conn *pgxpool.Conn, result any) error {
		_, err := conn.Exec(context.Background(),
			"insert into misc_small_object (obj_id, file_content) values ($1, $2);",
			docId, data)
		return err
	}); err != nil {
		log.Println("insert uploaded file error:", err)
		ctx.JSON(http.StatusInternalServerError, ResponseErr("5000", "internal error"))
		return
	}

	ctx.JSON(http.StatusOK, ResponseOk())
}

func DeleteSmallObject(ctx *gin.Context) {
	docId := ctx.Param("doc_id")

	if _, err := pgbackend.RunQuery(SmallObjectServiceName, nil, func(conn *pgxpool.Conn, result any) error {
		_, err := conn.Exec(context.Background(),
			"delete from misc_small_object where obj_id = $1;",
			docId)
		return err
	}); err != nil {
		log.Println("delete uploaded file error:", err)
		ctx.JSON(http.StatusInternalServerError, ResponseErr("5000", "internal error"))
		return
	}

	ctx.JSON(http.StatusOK, ResponseOk())
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

func GetUploadedFile(file *multipart.FileHeader) ([]byte, error) {
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	buf := new(bytes.Buffer)

	_, err = io.Copy(buf, src)
	return buf.Bytes(), err
}
