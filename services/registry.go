package services

import (
	"payment-service/repositories"
)

type Registry struct {
	repository repositories.IRepositoryRegistry
}

type IServiceRegistry interface {
}

func NewServiceRegistry(repository repositories.IRepositoryRegistry) IServiceRegistry {
	return &Registry{
		repository: repository,
	}
}
