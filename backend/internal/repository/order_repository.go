package repository

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/jackc/pgx/v5"

	"mini-erp-test/backend/internal/apperror"
	"mini-erp-test/backend/internal/model"
)

type lockedProduct struct {
	id int
	price int
	stock int
	isActive bool
}

func (r *Repository) CreateOrder(ctx context.Context, in model.OrderInput) (model.Order, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return model.Order{}, err
	}
	defer tx.Rollback(ctx)
	skus := make([]string, len(in.Items))
	for i, item := range in.Items {
		skus[i] = item.SKU
	}
	rows, err := tx.Query(ctx,
		`SELECT id, sku, price, stock, is_active
		 FROM products
		 WHERE sku = ANY($1)
		 ORDER BY id
		 FOR UPDATE`,
		skus,
	)
	if err != nil {
		return model.Order{}, err
	}
	products := map[string]lockedProduct{}
	for rows.Next() {
		var sku string
		var p lockedProduct
		if err := rows.Scan(&p.id, &sku, &p.price, &p.stock, &p.isActive); err != nil {
			rows.Close()
			return model.Order{}, err
		}
		products[sku] = p
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return model.Order{}, err
	}
	fields := map[string]string{}
	for i, item := range in.Items {
		p, found := products[item.SKU]
		if !found {
			fields[fmt.Sprintf("items[%d].sku", i)] = "Unknown SKU " + item.SKU
		} else if !p.isActive {
			fields[fmt.Sprintf("items[%d].sku", i)] = item.SKU + " is inactive and cannot be sold"
		}
	}
	if len(fields) > 0 {
		return model.Order{}, apperror.Validation(fields)
	}

	for _, item := range in.Items {
		p := products[item.SKU]
		if p.stock < item.Qty {
			return model.Order{}, apperror.InsufficientStock(item.SKU, p.stock)
		}
	}

	total := 0
	for _, item := range in.Items {
		total += products[item.SKU].price * item.Qty
	}
	if total > math.MaxInt32 {
		return model.Order{}, apperror.Validation(map[string]string{"items": "Order total is too large"})
	}

	var orderID int
	err = tx.QueryRow(ctx,
		`INSERT INTO sales_orders (customer, total) VALUES ($1, $2) RETURNING id`,
		in.Customer, total,
	).Scan(&orderID)
	if err != nil {
		return model.Order{}, err
	}

	note := fmt.Sprintf("Order #%d", orderID)
	for _, item := range in.Items {
		p := products[item.SKU]
		tag, err := tx.Exec(ctx,
			`UPDATE products SET stock = stock - $1 WHERE id = $2 AND stock >= $1`,
			item.Qty, p.id,
		)
		if err != nil {
			return model.Order{}, err
		}
		if tag.RowsAffected() == 0 {
			return model.Order{}, apperror.InsufficientStock(item.SKU, p.stock)
		}

		_, err = tx.Exec(ctx,
			`INSERT INTO sales_order_items (order_id, sku, qty, unit_price) VALUES ($1, $2, $3, $4)`,
			orderID, item.SKU, item.Qty, p.price,
		)
		if err != nil {
			return model.Order{}, err
		}

		_, err = tx.Exec(ctx,
			`INSERT INTO stock_movements (product_id, type, qty, note) VALUES ($1, 'out', $2, $3)`,
			p.id, item.Qty, note,
		)
		if err != nil {
			return model.Order{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Order{}, err
	}
	return r.GetOrder(ctx, orderID)
}

func (r *Repository) GetOrder(ctx context.Context, id int) (model.Order, error) {
	var o model.Order
	err := r.db.QueryRow(ctx,
		`SELECT id, customer, status, total, created_at, cancelled_at FROM sales_orders WHERE id = $1`, id,
	).Scan(&o.ID, &o.Customer, &o.Status, &o.Total, &o.CreatedAt, &o.CancelledAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Order{}, apperror.NotFound("Order not found")
	}
	if err != nil {
		return model.Order{}, err
	}

	rows, err := r.db.Query(ctx,
		`SELECT i.sku, p.name, i.qty, i.unit_price
		 FROM sales_order_items i
		 JOIN products p ON p.sku = i.sku
		 WHERE i.order_id = $1
		 ORDER BY i.id`, id,
	)
	if err != nil {
		return model.Order{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var item model.OrderItem
		if err := rows.Scan(&item.SKU, &item.Name, &item.Qty, &item.UnitPrice); err != nil {
			return model.Order{}, err
		}
		o.Items = append(o.Items, item)
	}
	return o, rows.Err()
}

func (r *Repository) ListOrders(ctx context.Context) ([]model.Order, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, customer, status, total, created_at, cancelled_at
		 FROM sales_orders
		 ORDER BY id DESC
		 LIMIT 50`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := []model.Order{}
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(&o.ID, &o.Customer, &o.Status, &o.Total, &o.CreatedAt, &o.CancelledAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}

func (r *Repository) CancelOrder(ctx context.Context, id int) (model.Order, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return model.Order{}, err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx,
		`UPDATE sales_orders SET status = 'CANCELLED', cancelled_at = now()
		 WHERE id = $1 AND status = 'CREATED'`, id,
	)
	if err != nil {
		return model.Order{}, err
	}
	if tag.RowsAffected() == 0 {
		var exists bool
		err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM sales_orders WHERE id = $1)`, id).Scan(&exists)
		if err != nil {
			return model.Order{}, err
		}
		if !exists {
			return model.Order{}, apperror.NotFound("Order not found")
		}
		return model.Order{}, apperror.Conflict("ORDER_ALREADY_CANCELLED", "Order is already cancelled")
	}

	type restock struct{ productID, qty int }
	rows, err := tx.Query(ctx,
		`SELECT p.id, i.qty
		 FROM sales_order_items i
		 JOIN products p ON p.sku = i.sku
		 WHERE i.order_id = $1
		 ORDER BY p.id`, id,
	)
	if err != nil {
		return model.Order{}, err
	}
	var restocks []restock
	for rows.Next() {
		var rs restock
		if err := rows.Scan(&rs.productID, &rs.qty); err != nil {
			rows.Close()
			return model.Order{}, err
		}
		restocks = append(restocks, rs)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return model.Order{}, err
	}

	note := fmt.Sprintf("Order #%d cancelled", id)
	for _, rs := range restocks {
		if _, err := tx.Exec(ctx, `UPDATE products SET stock = stock + $1 WHERE id = $2`, rs.qty, rs.productID); err != nil {
			return model.Order{}, err
		}
		_, err := tx.Exec(ctx,
			`INSERT INTO stock_movements (product_id, type, qty, note) VALUES ($1, 'in', $2, $3)`,
			rs.productID, rs.qty, note,
		)
		if err != nil {
			return model.Order{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Order{}, err
	}
	return r.GetOrder(ctx, id)
}
