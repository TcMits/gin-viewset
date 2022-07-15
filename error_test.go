package viewset

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestViewSetNewViewSetError(t *testing.T) {
	baseErr := errors.New("testing base")
	err := NewViewSetError("testing", http.StatusBadRequest, baseErr)

	assert.Equal(t, "testing", err.Error())
	assert.Equal(t, "testing", err.message)
	assert.Equal(t, http.StatusBadRequest, err.StatusCode)
	assert.Equal(t, baseErr, err.ActualErr)
	assert.Equal(t, true, errors.Is(err, baseErr))
}

func TestDefaultExceptionHandlerHandleWithDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	handler := DefaultExceptionHandler{}
	blw := &MockGinBodyResponseWriter{MockBody: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw

	handler.Handle(errors.New("testing"), c)
	assert.Equal(t, `{"message":"testing"}`, blw.MockBody.String())
	assert.Equal(t, http.StatusBadRequest, blw.MockStatusCode)
}

func TestDefaultExceptionHandlerHandleWithViewSetError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	handler := DefaultExceptionHandler{}
	blw := &MockGinBodyResponseWriter{MockBody: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw
	baseErr := errors.New("testing")
	err := NewViewSetError("testing", http.StatusForbidden, baseErr)

	handler.Handle(*err, c)
	assert.Equal(t, `{"message":"testing"}`, blw.MockBody.String())
	assert.Equal(t, http.StatusForbidden, blw.MockStatusCode)
}

func TestDefaultExceptionHandlerHandleWithPointerViewSetError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	handler := DefaultExceptionHandler{}
	blw := &MockGinBodyResponseWriter{MockBody: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw
	baseErr := errors.New("testing")
	err := NewViewSetError("testing", http.StatusForbidden, baseErr)

	handler.Handle(err, c)
	assert.Equal(t, `{"message":"testing"}`, blw.MockBody.String())
	assert.Equal(t, http.StatusForbidden, blw.MockStatusCode)
}
