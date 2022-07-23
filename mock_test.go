package viewset

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"

	"github.com/gin-gonic/gin"
)

type MockSerializerAlwaysError[EntityType any] struct{}

type MockDeniedAny struct{}

type MockGinBodyResponseWriter struct {
	gin.ResponseWriter
	MockBody       *bytes.Buffer
	MockStatusCode int
}

func (_ *MockDeniedAny) Check(_ string, _ *gin.Context) error {
	return errors.New("Denied")
}

func (s *MockSerializerAlwaysError[EntityType]) Serialize(
	_ *map[string]any, _ *EntityType, _ *gin.Context,
) error {
	return errors.New("Serialize error")
}

func (s *MockSerializerAlwaysError[EntityType]) ManySerialize(
	_ *[]map[string]any, _ *[]*EntityType, _ *gin.Context,
) error {
	return errors.New("Many serialize error")
}

func (w *MockGinBodyResponseWriter) Write(b []byte) (int, error) {
	w.MockBody.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *MockGinBodyResponseWriter) WriteHeader(statusCode int) {
	w.MockStatusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func MockJsonPost(c *gin.Context, content any) {
	c.Request.Method = "POST" // or PUT
	c.Request.Header.Set("Content-Type", "application/json")

	jsonbytes, err := json.Marshal(content)
	if err != nil {
		panic(err)
	}

	// the request body must be an io.ReadCloser
	// the bytes buffer though doesn't implement io.Closer,
	// so you wrap it in a no-op closer
	c.Request.Body = io.NopCloser(bytes.NewBuffer(jsonbytes))
}

func MockJsonPut(c *gin.Context, content any) {
	c.Request.Method = "PUT"
	c.Request.Header.Set("Content-Type", "application/json")

	jsonbytes, err := json.Marshal(content)
	if err != nil {
		panic(err)
	}

	// the request body must be an io.ReadCloser
	// the bytes buffer though doesn't implement io.Closer,
	// so you wrap it in a no-op closer
	c.Request.Body = io.NopCloser(bytes.NewBuffer(jsonbytes))
}
