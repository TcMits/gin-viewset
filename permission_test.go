package viewset

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAllowAnyCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	checker := AllowAny{}
	err := checker.Check(DEFAULT_CREATE_ACTION, c)

	assert.Equal(t, nil, err)
}
