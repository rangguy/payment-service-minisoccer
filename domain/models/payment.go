package models

import (
	"github.com/google/uuid"
	"payment-service/constants"
	"time"
)

type Payment struct {
	ID               uint                     `gorm:"primaryKey;autoIncrement" `
	UUID             uuid.UUID                `gorm:"type:uuid;not null"`
	Amount           float64                  `gorm:"not null"`
	Status           *constants.PaymentStatus `gorm:"not null"`
	PaymentLink      string                   `gorm:"type:varchar(255);not null"`
	InvoiceLink      *string                  `gorm:"type:varchar(255);not null"`
	VANumber         *string                  `gorm:"type:varchar(50);not null"`
	Bank             *string                  `gorm:"type:varchar(100);not null"`
	Acquirer         *string                  `gorm:"type:varchar(100);not null"`
	TransactionID    *string                  `gorm:"type:varchar(255);not null"`
	Description      *string                  `gorm:"type:text;not null"`
	PaidAt           *time.Time               `gorm:"type:timestamp"`
	ExpiredAt        time.Time                `gorm:"type:timestamp"`
	CreatedAt        time.Time                `gorm:"type:timestamp;not null"`
	UpdatedAt        time.Time                `gorm:"type:timestamp;not null"`
	PaymentHistories []PaymentHistory         `gorm:"foreignKey:payment_id;references:id;OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
