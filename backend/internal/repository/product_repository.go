package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"mini-erp-test/backend/internal/apperror"
	"mini-erp-test/backend/internal/model"
)

const selectProduct = `
	SELECT p.id, p.sku, p.name, p.price, p.stock, p.is_active, p.created_at,
	       (SELECT MAX(m.created_at) FROM stock_movements m WHERE m.product_id = p.id)
	FROM products p`

func scanProduct(row pgx.Row) (model.Product, error) {
	var p model.Product
	err := row.Scan(&p.ID, &p.SKU, &p.Name, &p.Price, &p.Stock, &p.IsActive, &p.CreatedAt, &p.LastMovementAt)
	return p, err
}

func (r *Repository) ListProducts(ctx context.Context, search, status string, page, limit int) (model.ProductList, error) {
	where := " WHERE (p.sku ILIKE $1 OR p.name ILIKE $1)"
	switch status {
	case "active":
		where += " AND p.is_active"
	case "inactive":
		where += " AND NOT p.is_active"
	case "low_stock":
		where += " AND p.stock <= 10"
	}
	pattern := "%" + search + "%"

	var total int
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM products p"+where, pattern).Scan(&total)
	if err != nil {
		return model.ProductList{}, err
	}

	offset := (page - 1) * limit
	rows, err := r.db.Query(ctx, selectProduct+where+" ORDER BY p.name LIMIT $2 OFFSET $3", pattern, limit, offset)
	if err != nil {
		return model.ProductList{}, err
	}
	defer rows.Close()

	products := []model.Product{}
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return model.ProductList{}, err
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return model.ProductList{}, err
	}

	return model.ProductList{Data: products, Page: page, Limit: limit, Total: total}, nil
}

func (r *Repository) GetProduct(ctx context.Context, id int) (model.Product, error) {
	p, err := scanProduct(r.db.QueryRow(ctx, selectProduct+" WHERE p.id = $1", id))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Product{}, apperror.NotFound("Product not found")
	}
	return p, err
}

func (r *Repository) CreateProduct(ctx context.Context, in model.ProductInput) (model.Product, error) {
	isActive := true
	if in.IsActive != nil {
		isActive = *in.IsActive
	}

	var id int
	err := r.db.QueryRow(ctx,
		`INSERT INTO products (sku, name, price, is_active) VALUES ($1, $2, $3, $4) RETURNING id`,
		in.SKU, in.Name, *in.Price, isActive,
	).Scan(&id)
	if pgErrorCode(err) == uniqueViolation {
		return model.Product{}, apperror.Conflict("SKU_ALREADY_EXISTS", "SKU "+in.SKU+" already exists")
	}
	if err != nil {
		return model.Product{}, err
	}
	return r.GetProduct(ctx, id)
}

func (r *Repository) UpdateProduct(ctx context.Context, id int, in model.ProductInput) (model.Product, error) {
	tag, err := r.db.Exec(ctx,
		`UPDATE products
		 SET sku = $1, name = $2, price = $3, is_active = COALESCE($4, is_active)
		 WHERE id = $5`,
		in.SKU, in.Name, *in.Price, in.IsActive, id,
	)
	switch pgErrorCode(err) {
	case uniqueViolation:
		return model.Product{}, apperror.Conflict("SKU_ALREADY_EXISTS", "SKU "+in.SKU+" already exists")
	case foreignKeyViolation:
		return model.Product{}, apperror.Conflict("PRODUCT_HAS_HISTORY", "SKU cannot be changed because it is used in orders")
	}
	if err != nil {
		return model.Product{}, err
	}
	if tag.RowsAffected() == 0 {
		return model.Product{}, apperror.NotFound("Product not found")
	}
	return r.GetProduct(ctx, id)
}

func (r *Repository) DeleteProduct(ctx context.Context, id int) error {
	tag, err := r.db.Exec(ctx,
		`DELETE FROM products
		 WHERE id = $1 AND NOT EXISTS (SELECT 1 FROM stock_movements WHERE product_id = $1)`,
		id,
	)
	if pgErrorCode(err) == foreignKeyViolation {
		return productHasHistory()
	}
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 1 {
		return nil
	}

	if _, err := r.GetProduct(ctx, id); err != nil {
		return err
	}
	return productHasHistory()
}

func productHasHistory() error {
	return apperror.Conflict("PRODUCT_HAS_HISTORY", "Product has stock history and cannot be deleted")
}

func (r *Repository) AdjustStock(ctx context.Context, id int, in model.StockInput) (model.Product, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return model.Product{}, err
	}
	defer tx.Rollback(ctx)

	var sql string
	if in.Type == "in" {
		sql = `UPDATE products SET stock = stock + $1 WHERE id = $2`
	} else {
		sql = `UPDATE products SET stock = stock - $1 WHERE id = $2 AND stock >= $1`
	}
	tag, err := tx.Exec(ctx, sql, in.Qty, id)
	if err != nil {
		return model.Product{}, err
	}

	if tag.RowsAffected() == 0 {
		var sku string
		var stock int
		err := tx.QueryRow(ctx, `SELECT sku, stock FROM products WHERE id = $1`, id).Scan(&sku, &stock)
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Product{}, apperror.NotFound("Product not found")
		}
		if err != nil {
			return model.Product{}, err
		}
		return model.Product{}, apperror.InsufficientStock(sku, stock)
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO stock_movements (product_id, type, qty, note) VALUES ($1, $2, $3, $4)`,
		id, in.Type, in.Qty, nullIfEmpty(in.Note),
	)
	if err != nil {
		return model.Product{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Product{}, err
	}
	return r.GetProduct(ctx, id)
}

func (r *Repository) ListMovements(ctx context.Context, productID int) ([]model.Movement, error) {
	if _, err := r.GetProduct(ctx, productID); err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx,
		`SELECT id, product_id, type, qty, note, created_at
		 FROM stock_movements
		 WHERE product_id = $1
		 ORDER BY created_at DESC, id DESC`,
		productID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	movements := []model.Movement{}
	for rows.Next() {
		var m model.Movement
		if err := rows.Scan(&m.ID, &m.ProductID, &m.Type, &m.Qty, &m.Note, &m.CreatedAt); err != nil {
			return nil, err
		}
		movements = append(movements, m)
	}
	return movements, rows.Err()
}

func (r *Repository) GetSummary(ctx context.Context) (model.Summary, error) {
	var s model.Summary
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*),
		       COUNT(*) FILTER (WHERE is_active),
		       COALESCE(SUM(price::BIGINT * stock), 0),
		       COUNT(*) FILTER (WHERE stock <= $1)
		FROM products`, model.LowStockLimit,
	).Scan(&s.TotalProducts, &s.ActiveProducts, &s.TotalStockValue, &s.LowStockCount)
	return s, err
}
