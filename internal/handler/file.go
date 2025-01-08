package handler

import (
	"kelarin/internal/service"
	"kelarin/internal/types"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-errors/errors"
)

type File interface {
	UploadImages(c *gin.Context)
}

type fileImpl struct {
	fileUploadSvc service.File
}

func NewFile(fileUploadSvc service.File) File {
	return &fileImpl{fileUploadSvc}
}

func (h fileImpl) UploadImages(c *gin.Context) {
	var req types.FileUploadImagesReq

	form, err := c.MultipartForm()
	if errors.Is(err, http.ErrNotMultipart) {
		c.Error(errors.New(types.AppErr{
			Code:    http.StatusBadRequest,
			Message: http.ErrNotMultipart.Error(),
		}))
		return
	} else if errors.Is(err, http.ErrMissingFile) {
		c.Error(errors.New(types.AppErr{
			Code:    http.StatusBadRequest,
			Message: "images is required",
		}))
		return
	} else if err != nil {
		c.Error(errors.New(err))
		return
	}

	files := form.File["images"]
	if len(files) == 0 {
		c.Error(errors.New(types.AppErr{
			Code:    http.StatusBadRequest,
			Message: "images is required",
		}))
		return
	}

	req.Files = files

	res, err := h.fileUploadSvc.StoreTemp(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.ApiResponse{
		Code: http.StatusOK,
		Data: res,
	})
}
