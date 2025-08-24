package repositories

import (
	"gorm.io/gorm"
	paymentRepository "payment-service/repositories/payment"
	paymentHistoryRepository "payment-service/repositories/payment_history"
)

type Registry struct {
	db *gorm.DB
}

type IRepositoryRegistry interface {
	GetPayment() paymentRepository.IPaymentRepository
	GetPaymentHistory() paymentHistoryRepository.IPaymentHistoryRepository
}

func NewRepositoryRegistry(db *gorm.DB) IRepositoryRegistry {
	return &Registry{
		db: db,
	}
}

func (p *Registry) GetPayment() paymentRepository.IPaymentRepository {
	return paymentRepository.NewPaymentRepository(p.db)
}

func (p *Registry) GetPaymentHistory() paymentHistoryRepository.IPaymentHistoryRepository {
	return paymentHistoryRepository.NewPaymentHistoryRepository(p.db)
}
