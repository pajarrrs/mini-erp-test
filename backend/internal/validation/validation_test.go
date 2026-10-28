package validation

import (
	"testing"

	"mini-erp-test/backend/internal/model"
)

func intPtr(n int) *int { return &n }

func TestProduct(t *testing.T) {
	valid := model.ProductInput{SKU: "RICE", Name: "Rice", Price: intPtr(0)}
	if errs := Product(valid); len(errs) != 0 {
		t.Errorf("valid product got errors: %v", errs)
	}

	errs := Product(model.ProductInput{Price: intPtr(-1)})
	for _, field := range []string{"sku", "name", "price"} {
		if errs[field] == "" {
			t.Errorf("expected an error for %q, got %v", field, errs)
		}
	}

	if errs := Product(model.ProductInput{SKU: "A", Name: "A"}); errs["price"] == "" {
		t.Errorf("missing price should be an error")
	}
}

func TestStock(t *testing.T) {
	if errs := Stock(model.StockInput{Type: "out", Qty: 3}); len(errs) != 0 {
		t.Errorf("valid input got errors: %v", errs)
	}
	errs := Stock(model.StockInput{Type: "sideways", Qty: 0})
	if errs["type"] == "" || errs["qty"] == "" {
		t.Errorf("expected type and qty errors, got %v", errs)
	}
}

func TestOrder(t *testing.T) {
	valid := model.OrderInput{Customer: "Toko A", Items: []model.OrderItemInput{{SKU: "RICE", Qty: 1}}}
	if errs := Order(valid); len(errs) != 0 {
		t.Errorf("valid order got errors: %v", errs)
	}

	tests := []struct {
		name string
		input model.OrderInput
		wantField string
	}{
		{"empty customer", model.OrderInput{Items: valid.Items}, "customer"},
		{"no items", model.OrderInput{Customer: "A"}, "items"},
		{"qty zero", model.OrderInput{Customer: "A", Items: []model.OrderItemInput{{SKU: "RICE", Qty: 0}}}, "items[0].qty"},
		{"qty negative", model.OrderInput{Customer: "A", Items: []model.OrderItemInput{{SKU: "RICE", Qty: -2}}}, "items[0].qty"},
		{"missing sku", model.OrderInput{Customer: "A", Items: []model.OrderItemInput{{SKU: "", Qty: 1}}}, "items[0].sku"},
		{"duplicate sku", model.OrderInput{Customer: "A", Items: []model.OrderItemInput{{SKU: "RICE", Qty: 1}, {SKU: "RICE", Qty: 2}}}, "items[1].sku"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := Order(tt.input)
			if errs[tt.wantField] == "" {
				t.Errorf("expected an error on %q, got %v", tt.wantField, errs)
			}
		})
	}
}
