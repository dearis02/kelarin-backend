package middleware

import (
	"kelarin/internal/types"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-errors/errors"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/rs/zerolog/log"
)

func HttpErrorHandler(c *gin.Context) {
	c.Next()

	if c.Errors.Last() == nil {
		return
	}

	err := c.Errors.Last().Unwrap()
	res := types.ApiResponse{
		StatusCode: http.StatusInternalServerError,
	}

	switch e := err.(type) {
	case *errors.Error:
		if appErr, ok := e.Err.(types.AppErr); ok {
			res.StatusCode = appErr.Code
			res.Message = appErr.Message
			if appErr.Err != nil {
				res.Message = appErr.Err.Error()
				log.Error().Stack().Err(appErr.Err).Send()
			}
		} else {
			log.Error().Stack().Err(err).Send()
		}
	case validation.Errors:
		res.StatusCode = http.StatusUnprocessableEntity
		res.Message = "Validation error"

		for key, val := range err.(validation.Errors) {
			res.Errors = append(res.Errors, types.ErrValidationRes{
				Field:   key,
				Message: val.Error(),
			})
		}
	default:
		log.Error().Stack().Err(err).Send()
	}

	if res.Message == "" {
		res.Message = http.StatusText(res.StatusCode)
	}

	c.JSON(res.StatusCode, res)
}
