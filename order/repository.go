package order

import (
	"context"
	"database/sql"
)

type Repository interface {
	Close()
	PutOrder(ctx context.Context, order Order) error
	GetOrdersForAccount(ctx context.Context, accountID string) ([]Order, error)
}

type postgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(url string) (Repository, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return &postgresRepository{
		db: db,
	}, nil
}

func (r *postgresRepository) Close() {
	r.db.Close()
}

func (r *postgresRepository) PutOrder(ctx context.Context, order Order) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()
	tx.ExecContext(
		ctx,
		"INSERT INTO orders (id, account_id, total_amount, created_at) VALUES ($1, $2, $3, $4)",
		order.ID, order.AccountID, order.TotalAmount, order.CreatedAt)

	if err != nil {
		return err
	}
	stmt, _ := tx.PrepareContext(
		ctx,
		"COPY order_products (order_id, product_id, quantity, price) FROM STDIN",
	)
	for _, p := range order.Products {
		_, err = stmt.ExecContext(ctx, order.ID, p.ID, p.Quantity, p.Price)
		if err != nil {
			return err
		}
	}
	_, err = stmt.ExecContext(ctx)
	if err != nil {
		return err
	}
	stmt.Close()
	return
}

func (r *postgresRepository) GetOrdersForAccount(ctx context.Context, accountID string) ([]Order, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT
		o.id, o.account_id, o.total_amount::money::numeric::float8, o.created_at,
		op.product_id, op.quantity, op.price::money::numeric::float8
		FROM orders o
		JOIN order_products op ON o.id = op.order_id
		WHERE o.account_id = $1
		ORDER BY o.created_at DESC`,
		accountID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := []Order{}
	order := &Order{}
	lastOrder := &Order{}
	orderedProduct := &OrderedProduct{}
	products := []OrderedProduct{}

	for rows.Next() {
		if err := rows.Scan(
			&order.ID,
			&order.AccountID,
			&order.CreatedAt,
			&order.TotalAmount,
			&orderedProduct.ID,
			&orderedProduct.Quantity,
			&orderedProduct.Price,
		); err != nil {
			return nil, err
		}
		if lastOrder.ID != "" && lastOrder.ID != order.ID {
			newOrder := Order{
				ID:          lastOrder.ID,
				AccountID:   lastOrder.AccountID,
				CreatedAt:   lastOrder.CreatedAt,
				TotalAmount: lastOrder.TotalAmount,
				Products:    products,
			}
			orders = append(orders, newOrder)
			products = []OrderedProduct{}
		}
		products = append(products, OrderedProduct{
			ID:       orderedProduct.ID,
			Quantity: orderedProduct.Quantity,
			Price:    orderedProduct.Price,
		})
		*lastOrder = *order
	}
	if lastOrder.ID != "" {
		newOrder := Order{
			ID:          lastOrder.ID,
			AccountID:   lastOrder.AccountID,
			CreatedAt:   lastOrder.CreatedAt,
			TotalAmount: lastOrder.TotalAmount,
			Products:    products,
		}
		orders = append(orders, newOrder)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}
