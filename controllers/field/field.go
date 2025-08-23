package controllers

import (
	errValidation "field-service/common/error"
	"field-service/common/response"
	"field-service/domain/dto"
	fieldService "field-service/services"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"net/http"
)

type FieldController struct {
	service fieldService.IServiceRegistry
}

type IFieldController interface {
	GetAllWithPagination(*gin.Context)
	GetAllWithoutPagination(*gin.Context)
	GetByUUID(*gin.Context)
	Create(*gin.Context)
	Update(*gin.Context)
	Delete(*gin.Context)
}

func NewFieldController(service fieldService.IServiceRegistry) IFieldController {
	return &FieldController{service: service}
}

func (f *FieldController) GetAllWithPagination(context *gin.Context) {
	var params dto.FieldRequestParam
	if err := context.ShouldBindQuery(&params); err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  context,
		})
		return
	}

	validate := validator.New()
	err := validate.Struct(params)
	if err != nil {
		errMessage := http.StatusText(http.StatusUnprocessableEntity)
		errorResponse := errValidation.ErrValidationResponse(err)

		response.HttpResponse(response.ParamHTTPResp{
			Code:    http.StatusUnprocessableEntity,
			Err:     err,
			Message: &errMessage,
			Data:    errorResponse,
			Gin:     context,
		})
		return
	}

	result, err := f.service.GetField().GetAllWithPagination(context, &params)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  context,
		})
		return
	}

	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Data: result,
		Gin:  context,
	})
}

func (f *FieldController) GetAllWithoutPagination(context *gin.Context) {
	result, err := f.service.GetField().GetAllWithoutPagination(context)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  context,
		})
		return
	}

	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Data: result,
		Gin:  context,
	})
}

func (f *FieldController) GetByUUID(context *gin.Context) {
	result, err := f.service.GetField().GetByUUID(context, context.Param("uuid"))
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  context,
		})
		return
	}

	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Data: result,
		Gin:  context,
	})
}

func (f *FieldController) Create(context *gin.Context) {
	var request dto.FieldRequest
	if err := context.ShouldBindWith(&request, binding.FormMultipart); err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  context,
		})
		return
	}

	validate := validator.New()
	err := validate.Struct(request)
	if err != nil {
		errMessage := http.StatusText(http.StatusUnprocessableEntity)
		errorResponse := errValidation.ErrValidationResponse(err)

		response.HttpResponse(response.ParamHTTPResp{
			Code:    http.StatusUnprocessableEntity,
			Err:     err,
			Message: &errMessage,
			Data:    errorResponse,
			Gin:     context,
		})
		return
	}

	result, err := f.service.GetField().Create(context, &request)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  context,
		})
		return
	}

	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Data: result,
		Gin:  context,
	})
}

func (f *FieldController) Update(context *gin.Context) {
	var request dto.UpdateFieldRequest
	if err := context.ShouldBindWith(&request, binding.FormMultipart); err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  context,
		})
		return
	}

	validate := validator.New()
	err := validate.Struct(request)
	if err != nil {
		errMessage := http.StatusText(http.StatusUnprocessableEntity)
		errorResponse := errValidation.ErrValidationResponse(err)

		response.HttpResponse(response.ParamHTTPResp{
			Code:    http.StatusUnprocessableEntity,
			Err:     err,
			Message: &errMessage,
			Data:    errorResponse,
			Gin:     context,
		})
		return
	}

	result, err := f.service.GetField().Update(context, context.Param("uuid"), &request)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  context,
		})
		return
	}

	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Data: result,
		Gin:  context,
	})
}

func (f *FieldController) Delete(context *gin.Context) {
	err := f.service.GetField().Delete(context, context.Param("uuid"))
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  context,
		})
		return
	}

	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Gin:  context,
	})
}
