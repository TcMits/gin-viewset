package viewset

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var _ ExceptionHandler = &DefaultExceptionHandler{}

type ViewSetError struct {
	message    string
	StatusCode int
	ActualErr  error
}

func (err ViewSetError) Error() string {
	return err.message
}

func (err ViewSetError) Unwrap() error {
	return err.ActualErr
}

type DefaultExceptionHandler struct{}

func NewViewSetError(message string, statusCode int, baseError error) *ViewSetError {
	return &ViewSetError{
		message:    message,
		StatusCode: statusCode,
		ActualErr:  baseError,
	}
}

func (h *DefaultExceptionHandler) Handle(err error, c *gin.Context) {
	switch foundedErr := err.(type) {
	case *ViewSetError:
		c.AbortWithStatusJSON(
			foundedErr.StatusCode,
			map[string]any{
				"message": foundedErr.Error(),
			},
		)
	case ViewSetError:
		c.AbortWithStatusJSON(
			foundedErr.StatusCode,
			map[string]any{
				"message": foundedErr.Error(),
			},
		)
	default:
		c.AbortWithStatusJSON(
			http.StatusBadRequest,
			map[string]any{
				"message": foundedErr.Error(),
			},
		)
	}
}
