package middleware_test

import (
	"kelarin/internal/middleware"
	"kelarin/internal/types"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-errors/errors"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/stretchr/testify/assert"
)

func TestHttpErrorHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.HttpErrorHandler)

	tests := []struct {
		name       string
		setup      func(c *gin.Context)
		expectCode int
		expectMsg  string
	}{
		{
			name: "AppErr error",
			setup: func(c *gin.Context) {
				appErr := types.AppErr{Code: http.StatusBadRequest, Message: "app error occurred"}
				c.Error(errors.New(appErr))
			},
			expectCode: http.StatusBadRequest,
			expectMsg:  "app error occurred",
		},
		{
			name: "Validation error",
			setup: func(c *gin.Context) {
				validationErr := validation.Errors{"field": errors.New("validation failed")}
				c.Error(validationErr)
			},
			expectCode: http.StatusUnprocessableEntity,
			expectMsg:  "Validation error",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			test.setup(c)

			middleware.HttpErrorHandler(c)

			assert.Equal(t, test.expectCode, w.Code)
			assert.Contains(t, w.Body.String(), test.expectMsg)
		})
	}
}
