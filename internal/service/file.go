package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"kelarin/internal/config"
	"kelarin/internal/queue/task"
	"kelarin/internal/repository"
	"kelarin/internal/types"
	pkg "kelarin/pkg/utils"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gabriel-vasile/mimetype"
	"github.com/go-errors/errors"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type File interface {
	StoreTemp(ctx context.Context, req types.FileUploadImagesReq) (types.FileUploadFilesRes, error)
	GetTemp(ctx context.Context, fileName string) (types.FileGetTempRes, error)
	DeleteTemp(ctx context.Context, fileName string) error
	BulkUploadToS3(ctx context.Context, req []types.FileTemp, dir string) ([]string, error)
	GetS3PresignedURL(ctx context.Context, objectKey string) (string, error)
}

type fileImpl struct {
	redis           *redis.Client
	cfg             *config.Config
	fileRepo        repository.File
	tempFileTask    task.TempFile
	s3PresignClient *s3.PresignClient
	s3Uploader      *manager.Uploader
	s3Client        *s3.Client
}

func NewFile(redis *redis.Client, cfg *config.Config, fileRepo repository.File, tempFileTask task.TempFile, s3PresignClient *s3.PresignClient, s3Uploader *manager.Uploader, s3Client *s3.Client) File {
	return &fileImpl{
		redis,
		cfg,
		fileRepo,
		tempFileTask,
		s3PresignClient,
		s3Uploader,
		s3Client,
	}
}

func (r *fileImpl) StoreTemp(ctx context.Context, req types.FileUploadImagesReq) (types.FileUploadFilesRes, error) {
	var err error
	res := types.FileUploadFilesRes{}

	storedFiles := []types.FileTemp{}
	expiration := time.Until(time.Now().Add(r.cfg.File.TempFileExpiration))

	defer func() {
		if err != nil && len(storedFiles) > 0 {
			for _, file := range storedFiles {
				err := os.Remove(filepath.Join(types.TempFileDir, file.Name))
				if err != nil {
					log.Error().Err(fmt.Errorf("failed to remove temp file: %s", file.Name)).Send()
				}
			}
		}
	}()

	allowedMimeTypesMap := map[string]struct{}{}
	for _, mimeType := range r.cfg.File.AllowedImageTypes {
		allowedMimeTypesMap[mimeType] = struct{}{}
	}

	for _, file := range req.Files {
		if file.Size > r.cfg.File.UploadedImageFileSizeLimit {
			return res, errors.New(types.AppErr{
				Code:    http.StatusRequestEntityTooLarge,
				Message: fmt.Sprintf("file size is too large, allowed max file size %s", r.cfg.File.MaxUploadedImageFileSize),
			})
		}

		src, err := file.Open()
		if err != nil {
			return res, errors.New(err)
		}
		defer src.Close()

		srcBinary, err := io.ReadAll(src)
		if err != nil {
			return res, errors.New(err)
		}

		reader := bytes.NewReader(srcBinary)
		mimeType, err := mimetype.DetectReader(reader)
		if err != nil {
			return res, errors.New(err)
		}

		if _, ok := allowedMimeTypesMap[mimeType.String()]; !ok {
			return res, errors.New(types.AppErr{
				Code:    http.StatusUnsupportedMediaType,
				Message: fmt.Sprintf("file type %s is not allowed", mimeType.Extension()),
			})
		}

		fileName := pkg.GenerateUniqueFileName()
		filePath := filepath.Join(types.TempFileDir, fileName)
		err = os.WriteFile(filePath, srcBinary, 0644)
		if err != nil {
			return res, errors.New(err)
		}

		fileMeta := types.FileTemp{
			Name:       fileName,
			Expiration: expiration,
		}

		// add to queue to delete temp file after expiration
		defaultQueueName := types.GetQueueName(types.QueuePriorityDefault, r.cfg.Environment)
		err = r.tempFileTask.Delete(ctx, defaultQueueName, fileMeta, expiration)
		if err != nil {
			return res, errors.New(err)
		}

		storedFiles = append(storedFiles, fileMeta)
		res.FileKeys = append(res.FileKeys, fileName)
	}

	err = r.fileRepo.SetTemp(ctx, storedFiles)
	if err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (r *fileImpl) GetTemp(ctx context.Context, fileName string) (types.FileGetTempRes, error) {
	res := types.FileGetTempRes{}

	tempFile, err := r.fileRepo.GetTemp(ctx, fileName)
	if errors.Is(err, types.ErrNoData) {
		return res, errors.New(types.AppErr{
			Code:    http.StatusNotFound,
			Message: fmt.Sprintf("file %s not found", fileName),
		})
	} else if err != nil {
		return res, errors.New(err)
	}

	return types.FileGetTempRes(tempFile), nil
}

func (r *fileImpl) DeleteTemp(ctx context.Context, fileName string) error {
	if err := r.fileRepo.DeleteTemp(ctx, fileName); err != nil {
		return err
	}

	os.Remove(filepath.Join(types.TempFileDir, fileName))

	return nil
}

func (r *fileImpl) BulkUploadToS3(ctx context.Context, req []types.FileTemp, dir string) ([]string, error) {
	res := []string{}

	var err error

	defer func() {
		if err != nil && len(res) > 0 {
			for _, key := range res {
				r.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
					Bucket: aws.String(r.cfg.File.AwsS3Bucket),
					Key:    aws.String(key),
				})
			}
		}
	}()

	for _, file := range req {
		dest := filepath.Join(dir, file.Name)
		tempFilePath := filepath.Join(types.TempFileDir, file.Name)

		fileBinary, err := os.Open(tempFilePath)
		if errors.Is(err, os.ErrNotExist) {
			return res, errors.New(types.AppErr{
				Code:    http.StatusNotFound,
				Message: fmt.Sprintf("file %s not found", file.Name),
			})
		} else if err != nil {
			return res, errors.New(err)
		}

		defer fileBinary.Close()

		uploadRes, err := r.s3Uploader.Upload(ctx, &s3.PutObjectInput{
			Bucket: aws.String(r.cfg.File.AwsS3Bucket),
			Key:    aws.String(dest),
			Body:   fileBinary,
		})
		if err != nil {
			return res, errors.New(err)
		}

		res = append(res, *uploadRes.Key)
	}

	return res, nil
}

func (r *fileImpl) GetS3PresignedURL(ctx context.Context, objectKey string) (string, error) {
	req := &s3.GetObjectInput{
		Bucket: &r.cfg.File.AwsS3Bucket,
		Key:    &objectKey,
	}

	signedRes, err := r.s3PresignClient.PresignGetObject(ctx, req)
	if err != nil {
		return "", err
	}

	return signedRes.URL, nil
}
