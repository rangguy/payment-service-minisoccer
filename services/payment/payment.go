package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"math/rand"
	"os"
	"path/filepath"
	clients "payment-service/clients/midtrans"
	"payment-service/common/util"
	"payment-service/config"
	"payment-service/constants"
	errPayment "payment-service/constants/error/payment"
	"payment-service/controllers/kafka"
	"payment-service/domain/dto"
	"payment-service/domain/models"
	"payment-service/repositories"
	"regexp"
	"strings"
	"time"
)

type PaymentService struct {
	repository repositories.IRepositoryRegistry
	kafka      kafka.IKafkaRegistry
	midtrans   clients.IMidtransClient
}

type IPaymentService interface {
	GetAllWithPagination(context.Context, *dto.PaymentRequestParam) (*util.PaginationResult, error)
	GetByUUID(context.Context, string) (*dto.PaymentResponse, error)
	Create(context.Context, *dto.PaymentRequest) (*dto.PaymentResponse, error)
	Webhook(context.Context, *dto.WebHook) error
}

func NewPaymentService(repository repositories.IRepositoryRegistry) *PaymentService {
	return &PaymentService{repository: repository}
}

func (p *PaymentService) GetAllWithPagination(ctx context.Context, param *dto.PaymentRequestParam) (*util.PaginationResult, error) {
	payments, total, err := p.repository.GetPayment().FindAllWithPagination(ctx, param)
	if err != nil {
		return nil, err
	}

	paymentResults := make([]*dto.PaymentResponse, 0, len(payments))
	for _, payment := range payments {
		paymentResults = append(paymentResults, &dto.PaymentResponse{
			UUID:          payment.UUID,
			TransactionID: payment.TransactionID,
			OrderID:       payment.OrderID,
			Amount:        payment.Amount,
			Status:        payment.Status.GetStatusString(),
			PaymentLink:   payment.PaymentLink,
			InvoiceLink:   payment.InvoiceLink,
			VANumber:      payment.VANumber,
			Bank:          payment.Bank,
			Description:   payment.Description,
			ExpiredAt:     payment.ExpiredAt,
			CreatedAt:     payment.CreatedAt,
			UpdatedAt:     payment.UpdatedAt,
		})
	}

	paginationParam := util.PaginationParam{
		Page:  param.Page,
		Limit: param.Limit,
		Count: total,
		Data:  paymentResults,
	}

	response := util.GeneratePagination(paginationParam)
	return &response, nil
}

func (p *PaymentService) GetByUUID(ctx context.Context, uuid string) (*dto.PaymentResponse, error) {
	payment, err := p.repository.GetPayment().FindByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	return &dto.PaymentResponse{
		UUID:          payment.UUID,
		TransactionID: payment.TransactionID,
		OrderID:       payment.OrderID,
		Amount:        payment.Amount,
		Status:        payment.Status.GetStatusString(),
		PaymentLink:   payment.PaymentLink,
		InvoiceLink:   payment.InvoiceLink,
		VANumber:      payment.VANumber,
		Bank:          payment.Bank,
		Description:   payment.Description,
		ExpiredAt:     payment.ExpiredAt,
		CreatedAt:     payment.CreatedAt,
		UpdatedAt:     payment.UpdatedAt,
	}, nil

}

func (p *PaymentService) Create(ctx context.Context, request *dto.PaymentRequest) (*dto.PaymentResponse, error) {
	var (
		txErr, err error
		payment    *models.Payment
		response   *dto.PaymentResponse
		midtrans   *clients.MidtransData
	)

	err = p.repository.GetTx().Transaction(func(tx *gorm.DB) error {
		if !request.ExpiredAt.After(time.Now()) {
			return errPayment.ErrExpireAtInvalid
		}

		midtrans, txErr = p.midtrans.CreatePaymentLink(request)
		if txErr != nil {
			return txErr
		}

		paymentRequest := &dto.PaymentRequest{
			OrderID:     request.OrderID,
			Amount:      request.Amount,
			Description: request.Description,
			ExpiredAt:   request.ExpiredAt,
			PaymentLink: midtrans.RedirectURL,
		}

		payment, txErr = p.repository.GetPayment().Create(ctx, tx, paymentRequest)
		if txErr != nil {
			return txErr
		}

		txErr = p.repository.GetPaymentHistory().Create(ctx, tx, &dto.PaymentHistoryRequest{
			PaymentID: payment.ID,
			Status:    payment.Status.GetStatusString(),
		})
		return nil
	})

	if err != nil {
		return nil, err
	}

	response = &dto.PaymentResponse{
		UUID:        payment.UUID,
		OrderID:     payment.OrderID,
		Amount:      payment.Amount,
		Status:      payment.Status.GetStatusString(),
		PaymentLink: payment.PaymentLink,
		Description: payment.Description,
	}

	return response, nil
}

func (p *PaymentService) convertToIndonesianMonth(englishMonth string) string {
	monthMap := map[string]string{
		"January":   "Januari",
		"February":  "Februari",
		"March":     "Maret",
		"April":     "April",
		"May":       "Mei",
		"June":      "Juni",
		"July":      "Juli",
		"August":    "Agustus",
		"September": "September",
		"October":   "Oktober",
		"November":  "November",
		"December":  "Desember",
	}

	indonesianMonth, ok := monthMap[englishMonth]
	if !ok {
		return errors.New("month not found").Error()
	}

	return indonesianMonth
}

func (p *PaymentService) generatePDF(request *dto.InvoiceRequest) ([]byte, error) {
	htmlTemplatePath := "template/invoice.html"
	htmlTemplate, err := os.ReadFile(htmlTemplatePath)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	jsonData, _ := json.Marshal(request)
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		return nil, err
	}

	pdf, err := util.GeneratePDFFromHTML(string(htmlTemplate), data)
	if err != nil {
		return nil, err
	}

	return pdf, nil
}

func (p *PaymentService) uploadFile(ctx context.Context, invoiceNumber string, pdf []byte) (string, error) {
	if len(pdf) == 0 {
		return "", errors.New("pdf kosong")
	}

	baseDir := "invoice"
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return "", fmt.Errorf("gagal membuat folder %q: %w", baseDir, err)
	}

	clean := strings.ToLower(invoiceNumber)
	clean = strings.ReplaceAll(clean, "/", "")
	clean = strings.ReplaceAll(clean, `\`, "")
	re := regexp.MustCompile(`[^a-z0-9-_]+`)
	clean = strings.Trim(re.ReplaceAllString(clean, "-"), "-_")
	if clean == "" {
		clean = "invoice"
	}

	ts := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("%s-%s.pdf", clean, ts)

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	fullPath := filepath.Join(baseDir, filename)
	if err := os.WriteFile(fullPath, pdf, 0o644); err != nil {
		return "", fmt.Errorf("gagal menyimpan file: %w", err)
	}

	return filepath.ToSlash(fullPath), nil
}

func (p *PaymentService) randomNumber() int {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	number := random.Intn(900000) + 100000
	return number
}

func (p *PaymentService) mapTransactionStatusTOEvent(status constants.PaymentStatusString) string {
	var paymentStatus string
	switch status {
	case constants.PendingString:
		paymentStatus = strings.ToUpper(constants.PendingString.String())
	case constants.SettlementString:
		paymentStatus = strings.ToUpper(constants.SettlementString.String())
	case constants.ExpireString:
		paymentStatus = strings.ToUpper(constants.ExpireString.String())
	}
	return paymentStatus
}

func (p *PaymentService) produceToKafka(request *dto.WebHook, payment *models.Payment, paidAt *time.Time) error {
	event := dto.KafkaEvent{
		Name: p.mapTransactionStatusTOEvent(request.TransactionStatus),
	}

	metadata := dto.KafkaMetaData{
		Sender:    "payment-service",
		SendingAt: time.Now().Format(time.RFC3339),
	}

	body := dto.KafkaBody{
		Type: "JSON",
		Data: &dto.KafkaData{
			OrderID:   payment.OrderID,
			PaymentID: payment.UUID,
			Status:    request.TransactionStatus.String(),
			PaidAt:    paidAt,
		},
	}

	kafkaMessage := dto.KafkaMessage{
		Event:    event,
		Metadata: metadata,
		Body:     body,
	}

	topic := config.Config.Kafka.Topic
	kafkaMessageJSON, _ := json.Marshal(kafkaMessage)
	err := p.kafka.GetKafkaProducer().ProducerMessage(topic, kafkaMessageJSON)
	if err != nil {
		return err
	}

	return nil
}

func (p *PaymentService) Webhook(ctx context.Context, hook *dto.WebHook) error {
	
}
