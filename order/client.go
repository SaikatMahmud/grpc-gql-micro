package order

import (
	"context"
	"learn-grpc-gql-micro/order/pb"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn    *grpc.ClientConn
	service pb.OrderServiceClient
}

func NewClient(url string) (*Client, error) {
	conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	c := pb.NewOrderServiceClient(conn)
	return &Client{conn, c}, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) PostOrder(
	ctx context.Context,
	accountID string,
	products []OrderedProduct,
) (*Order, error) {
	protoProducts := []*pb.PostOrderRequest_OrderProduct{}
	for _, p := range products {
		protoProducts = append(protoProducts, &pb.PostOrderRequest_OrderProduct{
			ProductId: p.ID,
			Quantity:  uint32(p.Quantity),
		})
	}
	r, err := c.service.PostOrder(
		ctx,
		&pb.PostOrderRequest{
			AccountId: accountID,
			Products:  protoProducts,
		},
	)
	if err != nil {
		return nil, err
	}

	// Create response order
	newOrder := r.Order
	newOrderCreatedAt := time.Time{}
	newOrderCreatedAt.UnmarshalBinary(newOrder.CreatedAt)

	return &Order{
		ID:          newOrder.Id,
		CreatedAt:   newOrderCreatedAt,
		TotalAmount: newOrder.TotalAmount,
		AccountID:   newOrder.AccountId,
		Products:    products,
	}, nil
}

func (c *Client) GetOrdersForAccount(ctx context.Context, accountID string) ([]Order, error) {
	r, err := c.service.GetOrdersForAccount(ctx, &pb.GetOrdersForAccountRequest{
		AccountId: accountID,
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// Create response orders
	orders := []Order{}
	for _, orderProto := range r.Orders {
		newOrder := Order{
			ID:          orderProto.Id,
			TotalAmount: orderProto.TotalAmount,
			AccountID:   orderProto.AccountId,
		}
		newOrder.CreatedAt = time.Time{}
		newOrder.CreatedAt.UnmarshalBinary(orderProto.CreatedAt)

		products := []OrderedProduct{}
		for _, p := range orderProto.Products {
			products = append(products, OrderedProduct{
				ID:       p.Id,
				Quantity: p.Quantity,
				Price:    p.Price,
			})
		}
		newOrder.Products = products

		orders = append(orders, newOrder)
	}
	return orders, nil
}
