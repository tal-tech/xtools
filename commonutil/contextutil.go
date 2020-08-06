package commonutil

import (
	"github.com/gin-gonic/gin"
	"context"
)

func TransGinToGo(contextIns *gin.Context) context.Context {

	goCtx := context.Background()
	for k, v := range contextIns.Keys {
		goCtx = context.WithValue(goCtx, k, v)
	}
	return goCtx
}
