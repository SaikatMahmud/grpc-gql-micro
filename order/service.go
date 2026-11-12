package order

import (
	"context"
	"time"

	"github.com/segmentio/ksuid"
)

type Service interface {
	PostOrder(ctx context.Context, accountID string, products []OrderedProduct) (*Order, error)
	GetOrdersForAccount(ctx context.Context, accountID string) ([]Order, error)
}

type Order struct {
	ID          string           `json:"id"`
	AccountID   string           `json:"account_id"`
	TotalAmount float64          `json:"total_amount"`
	Products    []OrderedProduct `json:"products"`
	CreatedAt   time.Time        `json:"created_at"`
}

type OrderedProduct struct {
	ID       string  `json:"id"`
	Quantity uint32  `json:"quantity"`
	Price    float64 `json:"price"`
}

type orderService struct {
	repository Repository
}

func NewService(r Repository) Service {
	return &orderService{
		repository: r,
	}
}

func (s *orderService) PostOrder(ctx context.Context, accountID string, products []OrderedProduct) (*Order, error) {
	o := &Order{
		ID:        ksuid.New().String(),
		AccountID: accountID,
		Products:  products,
		CreatedAt: time.Now(),
	}
	o.TotalAmount = 0.0
	for _, p := range products {
		o.TotalAmount += p.Price * float64(p.Quantity)
	}

	err := s.repository.PutOrder(ctx, *o)
	if err != nil {
		return nil, err
	}
	return o, nil
}

func (s *orderService) GetOrdersForAccount(ctx context.Context, accountID string) ([]Order, error) {
	return s.repository.GetOrdersForAccount(ctx, accountID)
}
