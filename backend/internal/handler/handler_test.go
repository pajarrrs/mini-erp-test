package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"mini-erp-test/backend/internal/repository"
	"mini-erp-test/backend/internal/testutil"
)

func newServer(t *testing.T) *httptest.Server {
	srv := httptest.NewServer(New(repository.New(testutil.NewTestDB(t))).Routes())
	t.Cleanup(srv.Close)
	return srv
}

func call(t *testing.T, srv *httptest.Server, method, path, body string) (int, map[string]any) {
	t.Helper()
	req, _ := http.NewRequest(method, srv.URL+path, strings.NewReader(body))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var decoded map[string]any
	json.NewDecoder(res.Body).Decode(&decoded)
	return res.StatusCode, decoded
}

func TestOrderFlowOverHTTP(t *testing.T) {
	srv := newServer(t)

	status, product := call(t, srv, "POST", "/api/products", `{"sku":"RICE","name":"Beras","price":78500}`)
	if status != 201 {
		t.Fatalf("create product: status %d %v", status, product)
	}
	call(t, srv, "POST", "/api/products/1/stock", `{"type":"in","qty":3}`)

	status, body := call(t, srv, "POST", "/api/orders", `{"customer":"Toko A","items":[{"sku":"RICE","qty":5}]}`)
	errBody, _ := body["error"].(map[string]any)
	if status != 409 || errBody["code"] != "INSUFFICIENT_STOCK" || errBody["sku"] != "RICE" || errBody["available"] != 3.0 {
		t.Errorf("want 409 INSUFFICIENT_STOCK sku=RICE available=3, got %d %v", status, body)
	}

	status, body = call(t, srv, "POST", "/api/orders", `{"customer":"","items":[]}`)
	errBody, _ = body["error"].(map[string]any)
	fields, _ := errBody["fields"].(map[string]any)
	if status != 400 || fields["customer"] == nil || fields["items"] == nil {
		t.Errorf("want 400 with customer and items errors, got %d %v", status, body)
	}

	status, order := call(t, srv, "POST", "/api/orders", `{"customer":"Toko A","items":[{"sku":"RICE","qty":2,"price":1}]}`)
	if status != 201 || order["total"] != 157000.0 {
		t.Errorf("want 201 with total 157000, got %d %v", status, order)
	}

	status, body = call(t, srv, "DELETE", "/api/products/1", "")
	if status != 409 {
		t.Errorf("delete: want 409, got %d %v", status, body)
	}

	if status, _ := call(t, srv, "PUT", "/api/products/999", `{"sku":"X","name":"X","price":1}`); status != 404 {
		t.Errorf("update missing product: want 404, got %d", status)
	}
	if status, _ := call(t, srv, "GET", "/api/orders/999", ""); status != 404 {
		t.Errorf("get missing order: want 404, got %d", status)
	}
}
