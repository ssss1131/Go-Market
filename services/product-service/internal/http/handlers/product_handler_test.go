package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"GoProduct/internal/domain"
	"GoProduct/internal/http/handlers"
	"GoProduct/internal/http/middleware"
	"GoProduct/internal/repo"
	"GoProduct/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestServer(t *testing.T) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&domain.Product{}); err != nil {
		t.Fatalf("auto-migrate: %v", err)
	}

	productsRepo := repo.NewProducts(db)
	svc := service.NewProductService(productsRepo)
	h := handlers.NewProductHandler(svc)

	r := gin.New()

	// middleware-заглушка: просто кладёт user_id = 1
	authBypass := func(c *gin.Context) {
		c.Set(middleware.UserIDKey, uint(1))
		c.Next()
	}

	g := r.Group("/products")
	g.Use(authBypass)
	{
		g.POST("/", h.Create)
		g.GET("/", h.List)
		g.GET("/:id", h.Get)
		g.PUT("/:id", h.Update)
		g.DELETE("/:id", h.Delete)
	}

	return r
}

func TestProductHandler_CreateAndGet(t *testing.T) {
	r := setupTestServer(t)

	// --- Create ---
	body := map[string]interface{}{
		"name":        "Handler product",
		"description": "via handler",
		"price":       12.34,
	}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/products/", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("POST /products/ status = %d, want %d; body = %s", w.Code, http.StatusCreated, w.Body.String())
	}

	var created struct {
		ID          uint    `json:"id"`
		UserID      uint    `json:"user_id"`
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if created.UserID != 1 {
		t.Errorf("UserID = %d, want %d", created.UserID, 1)
	}

	// --- Get ---
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/products/"+strconv.Itoa(int(created.ID)), nil)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /products/:id status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestProductHandler_List(t *testing.T) {
	r := setupTestServer(t)

	// создаём 2 продукта
	for i := 0; i < 2; i++ {
		body := map[string]interface{}{
			"name":  "P" + strconv.Itoa(i),
			"price": float64(i+1) * 10,
		}
		b, _ := json.Marshal(body)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/products/", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("create product failed: %d %s", w.Code, w.Body.String())
		}
	}

	// проверяем список
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/products/", nil)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /products/ status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp) != 2 {
		t.Fatalf("len(products) = %d, want %d", len(resp), 2)
	}
}

func TestProductHandler_Update_NotFound(t *testing.T) {
	r := setupTestServer(t)

	body := map[string]interface{}{
		"name":        "New name",
		"description": "x",
		"price":       100,
	}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/products/9999", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("PUT /products/:id status = %d, want %d; body = %s", w.Code, http.StatusNotFound, w.Body.String())
	}
}

func TestProductHandler_Create_BadRequest(t *testing.T) {
	r := setupTestServer(t)

	// price <= 0 → сработает binding `gt=0`
	body := map[string]interface{}{
		"name":  "Bad product",
		"price": 0,
	}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/products/", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("POST /products/ status = %d, want %d; body = %s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

func TestProductHandler_Get_InvalidID(t *testing.T) {
	r := setupTestServer(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/products/abc", nil)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("GET /products/:id status = %d, want %d; body = %s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}
