package services

import (
	"bytes"
	"context"
	"field-service/common/util"
	errConstant "field-service/constants/error"
	"field-service/domain/dto"
	"field-service/domain/models"
	"field-service/repositories"
	"fmt"
	uuid2 "github.com/google/uuid"
	"io"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type FieldService struct {
	repository repositories.IRepositoryRegistry
}

type IFieldService interface {
	GetAllWithPagination(context.Context, *dto.FieldRequestParam) (*util.PaginationResult, error)
	GetAllWithoutPagination(context.Context) ([]dto.FieldResponse, error)
	GetByUUID(context.Context, string) (*dto.FieldResponse, error)
	Create(context.Context, *dto.FieldRequest) (*dto.FieldResponse, error)
	Update(context.Context, string, *dto.UpdateFieldRequest) (*dto.FieldResponse, error)
	Delete(context.Context, string) error
}

func NewFieldService(repository repositories.IRepositoryRegistry) IFieldService {
	return &FieldService{repository}
}

func (f *FieldService) GetAllWithPagination(ctx context.Context, param *dto.FieldRequestParam) (*util.PaginationResult, error) {
	fields, total, err := f.repository.GetField().FindAllWithPagination(ctx, param)
	if err != nil {
		return nil, err
	}

	fieldResult := make([]*dto.FieldResponse, 0, len(fields))
	for _, field := range fields {
		fieldResult = append(fieldResult, &dto.FieldResponse{
			UUID:         field.UUID,
			Code:         field.Code,
			Name:         field.Name,
			PricePerHour: field.PricePerHour,
			Images:       field.Images,
			CreatedAt:    field.CreatedAt,
			UpdatedAt:    field.UpdatedAt,
		})
	}

	pagination := &util.PaginationParam{
		Count: total,
		Page:  param.Page,
		Limit: param.Limit,
		Data:  fieldResult,
	}

	response := util.GeneratePagination(*pagination)
	return &response, nil
}

func (f *FieldService) GetAllWithoutPagination(ctx context.Context) ([]dto.FieldResponse, error) {
	fields, err := f.repository.GetField().FindAllWithoutPagination(ctx)
	if err != nil {
		return nil, err
	}

	fieldResult := make([]dto.FieldResponse, 0, len(fields))
	for _, field := range fields {
		fieldResult = append(fieldResult, dto.FieldResponse{
			UUID:         field.UUID,
			Code:         field.Code,
			Name:         field.Name,
			PricePerHour: field.PricePerHour,
			Images:       field.Images,
		})
	}

	return fieldResult, nil
}

func (f *FieldService) GetByUUID(ctx context.Context, uuid string) (*dto.FieldResponse, error) {
	field, err := f.repository.GetField().FindByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	fieldResult := &dto.FieldResponse{
		UUID:         field.UUID,
		Code:         field.Code,
		Name:         field.Name,
		PricePerHour: field.PricePerHour,
		Images:       field.Images,
		CreatedAt:    field.CreatedAt,
		UpdatedAt:    field.UpdatedAt,
	}

	return fieldResult, nil
}
func (f *FieldService) validateUpload(images []multipart.FileHeader) error {
	if images == nil || len(images) == 0 {
		return errConstant.ErrInvalidUploadFile
	}

	for _, image := range images {
		if image.Size > 5*1024*1024 {
			return errConstant.ErrSizeTooBig
		}
	}

	return nil
}

func (f *FieldService) processAndUploadImage(ctx context.Context, image multipart.FileHeader) (string, error) {
	file, err := image.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	buffer := new(bytes.Buffer)
	_, err = io.Copy(buffer, file)
	if err != nil {
		return "", err
	}

	timestamp := time.Now().Format("20060102150405")
	ext := path.Ext(image.Filename)
	newFilename := fmt.Sprintf("%s-%s%s", timestamp, strings.TrimSuffix(image.Filename, ext), ext)

	// Gunakan forward slash, bukan backslash
	outputPath := "images/" + newFilename

	err = os.MkdirAll("images", os.ModePerm)
	if err != nil {
		return "", err
	}

	// Untuk menyimpan file tetap gunakan filepath.Join (OS-specific)
	actualFilePath := filepath.Join("images", newFilename)
	err = os.WriteFile(actualFilePath, buffer.Bytes(), 0644)
	if err != nil {
		return "", err
	}

	// Return path dengan forward slash untuk database
	return outputPath, nil
}

func (f *FieldService) uploadImage(ctx context.Context, images []multipart.FileHeader) ([]string, error) {
	err := f.validateUpload(images)
	if err != nil {
		return nil, err
	}

	urls := make([]string, 0, len(images))
	for _, image := range images {
		url, err := f.processAndUploadImage(ctx, image)
		if err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}

	return urls, nil
}

func (f *FieldService) Create(ctx context.Context, request *dto.FieldRequest) (*dto.FieldResponse, error) {
	imageUrl, err := f.uploadImage(ctx, request.Images)
	if err != nil {
		return nil, err
	}

	field, err := f.repository.GetField().Create(ctx, &models.Field{
		Code:         request.Code,
		Name:         request.Name,
		Images:       imageUrl,
		PricePerHour: request.PricePerHour,
	})
	if err != nil {
		return nil, err
	}

	response := &dto.FieldResponse{
		UUID:         field.UUID,
		Code:         field.Code,
		Name:         field.Name,
		PricePerHour: field.PricePerHour,
		Images:       field.Images,
		CreatedAt:    field.CreatedAt,
		UpdatedAt:    field.UpdatedAt,
	}

	return response, nil
}

func (f *FieldService) Update(ctx context.Context, uuid string, request *dto.UpdateFieldRequest) (*dto.FieldResponse, error) {
	field, err := f.repository.GetField().FindByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	var imageUrl []string
	if request.Images != nil {
		imageUrl, err = f.uploadImage(ctx, request.Images)
		if err != nil {
			return nil, err
		}
	} else {
		imageUrl = field.Images
	}

	fieldUpdated, err := f.repository.GetField().Update(ctx, uuid, &models.Field{
		Code:         request.Code,
		Name:         request.Name,
		Images:       imageUrl,
		PricePerHour: request.PricePerHour,
	})
	if err != nil {
		return nil, err
	}

	uuidParsed, _ := uuid2.Parse(uuid)

	response := &dto.FieldResponse{
		UUID:         uuidParsed,
		Code:         fieldUpdated.Code,
		Name:         fieldUpdated.Name,
		PricePerHour: fieldUpdated.PricePerHour,
		Images:       fieldUpdated.Images,
		CreatedAt:    fieldUpdated.CreatedAt,
		UpdatedAt:    fieldUpdated.UpdatedAt,
	}

	return response, nil
}

func (f *FieldService) Delete(ctx context.Context, uuid string) error {
	_, err := f.repository.GetField().FindByUUID(ctx, uuid)
	if err != nil {
		return err
	}

	err = f.repository.GetField().Delete(ctx, uuid)
	if err != nil {
		return err
	}

	return nil
}
