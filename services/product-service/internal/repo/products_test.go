package repo

import (
	"testing"

	"GoProduct/internal/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&domain.Product{}); err != nil {
		t.Fatalf("auto-migrate: %v", err)
	}
	return db
}

func TestProducts_CreateAndGetByID(t *testing.T) {
	db := newTestDB(t)
	// без префикса repo, мы уже в этом пакете
	r := NewProducts(db)

	p := &domain.Product{
		UserID:      1,
		Name:        "Repo product",
		Description: "desc",
		Price:       5.5,
	}
	if err := r.Create(p); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if p.ID == 0 {
		t.Fatalf("expected non-zero ID after Create")
	}

	got, err := r.GetByID(p.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Name != p.Name || got.Price != p.Price || got.UserID != p.UserID {
		t.Errorf("GetByID() = %+v, want %+v", got, p)
	}
}

func TestProducts_List(t *testing.T) {
	db := newTestDB(t)
	r := NewProducts(db)

	_ = r.Create(&domain.Product{UserID: 1, Name: "A", Price: 1})
	_ = r.Create(&domain.Product{UserID: 1, Name: "B", Price: 2})

	list, err := r.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("len(List()) = %d, want %d", len(list), 2)
	}
}

func TestProducts_UpdateAndDelete(t *testing.T) {
	db := newTestDB(t)
	r := NewProducts(db)

	p := &domain.Product{UserID: 1, Name: "Old", Price: 1}
	if err := r.Create(p); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	p.Name = "New"
	p.Price = 10

	if err := r.Update(p); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, err := r.GetByID(p.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Name != "New" || got.Price != 10 {
		t.Errorf("Update() did not persist changes: got %+v", got)
	}

	if err := r.Delete(p.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if _, err := r.GetByID(p.ID); err == nil {
		t.Fatalf("expected error after Delete, got nil")
	}
}
