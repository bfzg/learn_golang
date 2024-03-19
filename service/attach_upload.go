package service

import (
	"fmt"
	"ginchat/common"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func Image(ctx *gin.Context) {
	w := ctx.Writer
	req := ctx.Request

	srcFile, head, err := req.FormFile("file")
	if err != nil {
		common.RespFail(w, err.Error())
	}

	//检查文件后缀

	suffix := ".png"
	ofilName := head.Filename
	tem := strings.Split(ofilName, ".")

	if len(tem) > 1 {
		suffix = "." + tem[len(tem)-1]
	}

	//报错文件
	fileName := fmt.Sprintln("%d%04d%s", time.Now().Unix(), rand.Int(), suffix)
	dstFile, err := os.Create("./asset/upload/" + fileName)
	if err != nil {
		common.RespFail(w, err.Error())
		return
	}
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		common.RespFail(w, err.Error())
	}
	url := "./asset/upload/" + fileName
	common.RespOk(w, url, "发送成功")
}
