package viewset

import "github.com/gin-gonic/gin"

var _ PermissionChecker = &AllowAny{}

type AllowAny struct{}

func (_ *AllowAny) Check(_ string, _ *gin.Context) error {
	return nil
}
