package repository

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"mini-erp-test/backend/internal/apperror"
	"mini-erp-test/backend/internal/model"
	"mini-erp-test/backend/internal/testutil"
)

var ctx = context.Background()

func setup(t *testing.T) *Repository {
	return New(testutil.NewTestDB(t))
}

func addProduct(t *testing.T, r *Repository, sku string, price, stock int) model.Product {
	t.Helper()
	p, err := r.CreateProduct(ctx, model.ProductInput{SKU: sku, Name: sku, Price: &price})
	if err != nil {
		t.Fatal(err)
	}
	if stock > 0 {
		p, err = r.AdjustStock(ctx, p.ID, model.StockInput{Type: "in", Qty: stock, Note: "test"})
		if err != nil {
			t.Fatal(err)
		}
	}
	return p
}

func count(t *testing.T, r *Repository, sql string) int {
	t.Helper()
	var n int
	if err := r.db.QueryRow(ctx, sql).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func stockOf(t *testing.T, r *Repository, id int) int {
	t.Helper()
	p, err := r.GetProduct(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	return p.Stock
}

func asAppError(t *testing.T, err error, wantCode string) *apperror.Error {
	t.Helper()
	var appErr *apperror.Error
	if !errors.As(err, &appErr) || appErr.Code != wantCode {
		t.Fatalf("want error %s, got %v", wantCode, err)
	}
	return appErr
}

func TestCreateOrder_InsufficientStock_WritesNothing(t *testing.T) {
	r := setup(t)
	rice := addProduct(t, r, "RICE", 78500, 10)
	eggs := addProduct(t, r, "EGGS", 56000, 2)
	movementsBefore := count(t, r, "SELECT COUNT(*) FROM stock_movements")

	_, err := r.CreateOrder(ctx, model.OrderInput{
		Customer: "Toko A",
		Items: []model.OrderItemInput{
			{SKU: "RICE", Qty: 5},
			{SKU: "EGGS", Qty: 3},
		},
	})

	appErr := asAppError(t, err, "INSUFFICIENT_STOCK")
	if appErr.SKU != "EGGS" || *appErr.Available != 2 {
		t.Errorf("want sku=EGGS available=2, got sku=%s available=%d", appErr.SKU, *appErr.Available)
	}
	if got := stockOf(t, r, rice.ID); got != 10 {
		t.Errorf("RICE stock changed to %d, want 10", got)
	}
	if got := stockOf(t, r, eggs.ID); got != 2 {
		t.Errorf("EGGS stock changed to %d, want 2", got)
	}
	if n := count(t, r, "SELECT COUNT(*) FROM sales_orders"); n != 0 {
		t.Errorf("%d orders saved, want 0", n)
	}
	if n := count(t, r, "SELECT COUNT(*) FROM sales_order_items"); n != 0 {
		t.Errorf("%d order items saved, want 0", n)
	}
	if n := count(t, r, "SELECT COUNT(*) FROM stock_movements"); n != movementsBefore {
		t.Errorf("movements changed from %d to %d", movementsBefore, n)
	}
}

func TestCreateOrder_HappyPath(t *testing.T) {
	r := setup(t)
	rice := addProduct(t, r, "RICE", 78500, 10)
	eggs := addProduct(t, r, "EGGS", 56000, 5)

	order, err := r.CreateOrder(ctx, model.OrderInput{
		Customer: "Toko A",
		Items: []model.OrderItemInput{{SKU: "RICE", Qty: 2}, {SKU: "EGGS", Qty: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}

	if order.Status != "CREATED" || len(order.Items) != 2 {
		t.Errorf("unexpected order: %+v", order)
	}
	if want := 2*78500 + 1*56000; order.Total != want {
		t.Errorf("total = %d, want %d", order.Total, want)
	}
	if order.Items[0].UnitPrice != 78500 {
		t.Errorf("unit price = %d, want 78500", order.Items[0].UnitPrice)
	}
	if got := stockOf(t, r, rice.ID); got != 8 {
		t.Errorf("RICE stock = %d, want 8", got)
	}
	if got := stockOf(t, r, eggs.ID); got != 4 {
		t.Errorf("EGGS stock = %d, want 4", got)
	}
	if n := count(t, r, "SELECT COUNT(*) FROM stock_movements WHERE type = 'out'"); n != 2 {
		t.Errorf("%d out movements, want 2", n)
	}
}

func TestCreateOrder_UnknownSKU(t *testing.T) {
	r := setup(t)
	addProduct(t, r, "RICE", 78500, 10)

	_, err := r.CreateOrder(ctx, model.OrderInput{
		Customer: "Toko A",
		Items: []model.OrderItemInput{{SKU: "RICE", Qty: 1}, {SKU: "NOPE", Qty: 1}},
	})
	appErr := asAppError(t, err, "VALIDATION_ERROR")
	if appErr.Fields["items[1].sku"] == "" {
		t.Errorf("want an error on items[1].sku, got %v", appErr.Fields)
	}
}

func TestAdjustStock(t *testing.T) {
	r := setup(t)
	soap := addProduct(t, r, "SOAP", 4500, 0)

	p, err := r.AdjustStock(ctx, soap.ID, model.StockInput{Type: "in", Qty: 5, Note: "delivery"})
	if err != nil {
		t.Fatal(err)
	}
	if p.Stock != 5 || p.LastMovementAt == nil {
		t.Errorf("after in: stock=%d last_movement_at=%v", p.Stock, p.LastMovementAt)
	}

	_, err = r.AdjustStock(ctx, soap.ID, model.StockInput{Type: "out", Qty: 9})
	appErr := asAppError(t, err, "INSUFFICIENT_STOCK")
	if *appErr.Available != 5 {
		t.Errorf("available = %d, want 5", *appErr.Available)
	}
	if n := count(t, r, "SELECT COUNT(*) FROM stock_movements"); n != 1 {
		t.Errorf("%d movements, want 1", n)
	}

	_, err = r.AdjustStock(ctx, 9999, model.StockInput{Type: "in", Qty: 1})
	asAppError(t, err, "NOT_FOUND")
}

func TestCancelOrder_RestoresStock(t *testing.T) {
	r := setup(t)
	rice := addProduct(t, r, "RICE", 78500, 10)
	order, err := r.CreateOrder(ctx, model.OrderInput{Customer: "A", Items: []model.OrderItemInput{{SKU: "RICE", Qty: 4}}})
	if err != nil {
		t.Fatal(err)
	}

	cancelled, err := r.CancelOrder(ctx, order.ID)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled.Status != "CANCELLED" || cancelled.CancelledAt == nil {
		t.Errorf("unexpected order after cancel: %+v", cancelled)
	}
	if got := stockOf(t, r, rice.ID); got != 10 {
		t.Errorf("stock after cancel = %d, want 10", got)
	}

	_, err = r.CancelOrder(ctx, order.ID)
	asAppError(t, err, "ORDER_ALREADY_CANCELLED")
	if got := stockOf(t, r, rice.ID); got != 10 {
		t.Errorf("stock after second cancel = %d, want 10", got)
	}
}

func TestDeleteProduct(t *testing.T) {
	r := setup(t)
	withHistory := addProduct(t, r, "OLD", 1000, 5)
	fresh := addProduct(t, r, "NEW", 1000, 0)

	asAppError(t, r.DeleteProduct(ctx, withHistory.ID), "PRODUCT_HAS_HISTORY")
	if err := r.DeleteProduct(ctx, fresh.ID); err != nil {
		t.Errorf("delete product without history: %v", err)
	}
	asAppError(t, r.DeleteProduct(ctx, fresh.ID), "NOT_FOUND")
}

func TestListProducts_SearchAndFilter(t *testing.T) {
	r := setup(t)
	addProduct(t, r, "BRS-5KG", 78500, 50)
	addProduct(t, r, "GULA-1KG", 17500, 3)

	list, err := r.ListProducts(ctx, "brs", "", 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if list.Total != 1 || list.Data[0].SKU != "BRS-5KG" {
		t.Errorf("search 'brs' got %+v", list)
	}

	list, err = r.ListProducts(ctx, "", "low_stock", 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if list.Total != 1 || list.Data[0].SKU != "GULA-1KG" {
		t.Errorf("low_stock filter got %+v", list)
	}
}

func TestConcurrentOrders_StockNeverNegative(t *testing.T) {
	r := setup(t)
	p := addProduct(t, r, "HOT", 10000, 10)

	const customers = 30
	var wg sync.WaitGroup
	results := make([]error, customers)
	for i := 0; i < customers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, results[i] = r.CreateOrder(ctx, model.OrderInput{
				Customer: fmt.Sprintf("Customer %d", i),
				Items: []model.OrderItemInput{{SKU: "HOT", Qty: 1}},
			})
		}(i)
	}
	wg.Wait()

	succeeded, rejected := 0, 0
	for _, err := range results {
		var appErr *apperror.Error
		switch {
		case err == nil:
			succeeded++
		case errors.As(err, &appErr) && appErr.Code == "INSUFFICIENT_STOCK":
			rejected++
		default:
			t.Errorf("unexpected error: %v", err)
		}
	}

	if succeeded != 10 || rejected != 20 {
		t.Errorf("succeeded=%d rejected=%d, want 10 and 20", succeeded, rejected)
	}
	if got := stockOf(t, r, p.ID); got != 0 {
		t.Errorf("final stock = %d, want 0", got)
	}
}

func TestConcurrentAdjustments_StockNeverNegative(t *testing.T) {
	r := setup(t)
	p := addProduct(t, r, "SOAP", 4500, 5)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.AdjustStock(ctx, p.ID, model.StockInput{Type: "out", Qty: 1})
		}()
	}
	wg.Wait()

	if got := stockOf(t, r, p.ID); got != 0 {
		t.Errorf("final stock = %d, want 0", got)
	}
	if n := count(t, r, "SELECT COUNT(*) FROM stock_movements WHERE type = 'out'"); n != 5 {
		t.Errorf("%d out movements, want 5", n)
	}
}
