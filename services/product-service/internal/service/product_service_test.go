package service_test

import (
	"testing"

	"GoProduct/internal/domain"
	"GoProduct/internal/repo"
	"GoProduct/internal/service"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestProductService(t *testing.T) *service.ProductService {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	// создаём таблицу products
	if err := db.AutoMigrate(&domain.Product{}); err != nil {
		t.Fatalf("auto-migrate: %v", err)
	}

	productsRepo := repo.NewProducts(db)
	return service.NewProductService(productsRepo)
}

func TestCreateAndGetProduct(t *testing.T) {
	svc := newTestProductService(t)

	created, err := svc.CreateProduct(service.CreateProductInput{
		UserID:      42,
		Name:        "Test product",
		Description: "some description",
		Price:       9.99,
	})
	if err != nil {
		t.Fatalf("CreateProduct() error = %v", err)
	}
	if created.ID == 0 {
		t.Fatalf("expected non-zero ID, got %d", created.ID)
	}

	got, err := svc.GetProduct(created.ID)
	if err != nil {
		t.Fatalf("GetProduct() error = %v", err)
	}

	if got.UserID != 42 {
		t.Errorf("UserID = %d, want %d", got.UserID, 42)
	}
	if got.Name != "Test product" {
		t.Errorf("Name = %q, want %q", got.Name, "Test product")
	}
	if got.Price != 9.99 {
		t.Errorf("Price = %v, want %v", got.Price, 9.99)
	}
}

func TestListProducts(t *testing.T) {
	svc := newTestProductService(t)

	_, _ = svc.CreateProduct(service.CreateProductInput{
		UserID: 1, Name: "P1", Price: 1,
	})
	_, _ = svc.CreateProduct(service.CreateProductInput{
		UserID: 1, Name: "P2", Price: 2,
	})

	list, err := svc.ListProducts()
	if err != nil {
		t.Fatalf("ListProducts() error = %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("len(ListProducts()) = %d, want %d", len(list), 2)
	}
}

func TestUpdateProduct(t *testing.T) {
	svc := newTestProductService(t)

	created, err := svc.CreateProduct(service.CreateProductInput{
		UserID:      1,
		Name:        "Old",
		Description: "old desc",
		Price:       5,
	})
	if err != nil {
		t.Fatalf("CreateProduct() error = %v", err)
	}

	updated, err := svc.UpdateProduct(service.UpdateProductInput{
		ID:          created.ID,
		Name:        "New",
		Description: "new desc",
		Price:       10,
	})
	if err != nil {
		t.Fatalf("UpdateProduct() error = %v", err)
	}

	if updated.Name != "New" || updated.Price != 10 {
		t.Errorf("UpdateProduct(): got %+v", updated)
	}
}

func TestDeleteProduct(t *testing.T) {
	svc := newTestProductService(t)

	created, err := svc.CreateProduct(service.CreateProductInput{
		UserID: 1,
		Name:   "ToDelete",
		Price:  3,
	})
	if err != nil {
		t.Fatalf("CreateProduct() error = %v", err)
	}

	if err := svc.DeleteProduct(created.ID); err != nil {
		t.Fatalf("DeleteProduct() error = %v", err)
	}

	_, err = svc.GetProduct(created.ID)
	if !service.IsNotFound(err) {
		t.Fatalf("expected IsNotFound error after delete, got %v", err)
	}
}

func TestGetProduct_NotFound(t *testing.T) {
	svc := newTestProductService(t)

	_, err := svc.GetProduct(99999)
	if !service.IsNotFound(err) {
		t.Fatalf("expected IsNotFound error, got %v", err)
	}
}
