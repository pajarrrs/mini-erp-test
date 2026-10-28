package validation

import (
	"fmt"

	"mini-erp-test/backend/internal/model"
)

const MaxQty = 1_000_000

const MaxPrice = 2_147_483_647

func qtyError(qty int) string {
	if qty <= 0 {
		return "Quantity must be greater than 0"
	}
	if qty > MaxQty {
		return fmt.Sprintf("Quantity must be at most %d", MaxQty)
	}
	return ""
}

func Product(in model.ProductInput) map[string]string {
	errs := map[string]string{}
	if in.SKU == "" {
		errs["sku"] = "SKU is required"
	}
	if in.Name == "" {
		errs["name"] = "Name is required"
	}
	if in.Price == nil {
		errs["price"] = "Price is required"
	} else if *in.Price < 0 {
		errs["price"] = "Price must be 0 or more"
	} else if *in.Price > MaxPrice {
		errs["price"] = "Price is too large"
	}
	return errs
}

func Stock(in model.StockInput) map[string]string {
	errs := map[string]string{}
	if in.Type != "in" && in.Type != "out" {
		errs["type"] = `Type must be "in" or "out"`
	}
	if msg := qtyError(in.Qty); msg != "" {
		errs["qty"] = msg
	}
	return errs
}

func Order(in model.OrderInput) map[string]string {
	errs := map[string]string{}
	if in.Customer == "" {
		errs["customer"] = "Customer is required"
	}
	if len(in.Items) == 0 {
		errs["items"] = "Order must have at least one item"
	}

	seen := map[string]bool{}
	for i, item := range in.Items {
		skuField := fmt.Sprintf("items[%d].sku", i)
		qtyField := fmt.Sprintf("items[%d].qty", i)

		if item.SKU == "" {
			errs[skuField] = "SKU is required"
		} else if seen[item.SKU] {
			errs[skuField] = "Duplicate SKU " + item.SKU + ", combine it into one line"
		}
		seen[item.SKU] = true

		if msg := qtyError(item.Qty); msg != "" {
			errs[qtyField] = msg
		}
	}
	return errs
}
