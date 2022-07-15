package viewset

import "github.com/gin-gonic/gin"

type AllowAny struct{}

func (_ *AllowAny) Check(_ string, _ *gin.Context) error {
	return nil
}
