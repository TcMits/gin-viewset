package viewset

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDefaultValidatorValidate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	c.Request = &http.Request{
		Header: make(http.Header),
	}

	MockJsonPost(c, map[string]any{"name": "test", "age": 1})

	validator := DefaultValidator[testObject, testObject]{}
	result := map[string]any{}

	err := validator.Validate(&result, nil, c)

	assert.Equal(t, err, nil)
	assert.Equal(t, 1, result["age"].(int))
	assert.Equal(t, "test", result["name"].(string))
}

func TestDefaultValidatorValidateWithError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	c.Request = &http.Request{
		Header: make(http.Header),
	}

	MockJsonPost(c, map[string]any{"name": "test"})

	validator := DefaultValidator[testObject, testObject]{}
	result := map[string]any{}

	err := validator.Validate(&result, nil, c)

	assert.NotEqual(t, nil, err)
}
