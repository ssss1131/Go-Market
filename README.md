# Go-Market

A microservices-based marketplace platform built with Go, featuring user authentication, product management, and shopping cart functionality.

## Architecture

The project consists of three microservices communicating via Kafka and backed by PostgreSQL:

| Service | Port | Description |
|---------|------|-------------|
| **user-service** | 8080 | User registration, authentication, email verification |
| **product-service** | 8081 | Product CRUD operations, shopping cart |
| **notifications-service** | - | Email notifications via Kafka consumer |

**Infrastructure:**
- **PostgreSQL 16** — Two databases: `go_market_user` and `go_market_product`
- **Apache Kafka 3.7** — Event streaming for async notifications
- **MailHog** — Local SMTP server for email testing (Web UI on port 8025)

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) (v20.10+)
- [Docker Compose](https://docs.docker.com/compose/install/) (v2.0+)

## Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/ssss1131/Go-Market.git
cd Go-Market
```

### 2. Configure Environment Variables

Copy the example environment file:

**macOS / Linux:**
```bash
cp .env.example .env
```

**Windows (Command Prompt):**
```cmd
copy .env.example .env
```

**Windows (PowerShell):**
```powershell
Copy-Item .env.example .env
```

The default `.env` values work out of the box for local development:

```env
PG_USERNAME=postgres
PG_PASSWORD=postgres
PG_PORT_HOST=5433

USER_PG_URL=postgres://postgres:postgres@postgres:5432/go_market_user?sslmode=disable
PRODUCT_PG_URL=postgres://postgres:postgres@postgres:5432/go_market_product?sslmode=disable

JWT_SECRET=some_super_mega_secret

KAFKA_BROKERS=kafka:9092
SMTP_HOST=mailhog
SMTP_PORT=1025
SMTP_USER=
SMTP_PASS=
SMTP_FROM="GoMarket <noreply@gomarket.local>"
BASE_URL=http://localhost:8080
```

### 3. Start All Services

**macOS / Linux / Windows:**
```bash
docker-compose up --build
```

Or run in detached mode:
```bash
docker-compose up --build -d
```

Wait for all services to be healthy (first startup may take 1-2 minutes while images build).

### 4. Verify Services Are Running

- **User Service:** http://localhost:8080
- **Product Service:** http://localhost:8081
- **MailHog UI:** http://localhost:8025

### 5. Stop Services

```bash
docker-compose down
```

To also remove volumes (database data):
```bash
docker-compose down -v
```

---

## Postman Collection

Import the provided collection and environment for easy API testing:

1. **Import Collection:** `Go-Market.postman_collection.json`
2. **Import Environment:** `Go-Market.postman_environment.json`
3. **Select Environment:** Choose "Go-Market Local" from the environment dropdown

The collection includes:
- All API endpoints with example requests
- Test scenarios (valid inputs, validation errors, auth failures)
- **Auto-save token:** Login requests automatically save the access token
- Pre-configured authorization headers

---

## User Registration Flow

Follow these steps to register and authenticate:

### Step 1: Register a New User
You can do it with postman collection in repo or like below
```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John",
    "surname": "Doe",
    "email": "john@example.com",
    "password": "securepass123"
  }'
```

Response:
```json
{
  "user_id": 1,
  "status": "PENDING",
  "message": "check your email to verify account"
}
```

### Step 2: Verify Email via MailHog

1. Open **MailHog** in your browser: http://localhost:8025
2. Find the verification email sent to your registered address
3. Click the verification link or copy the token from the URL
4. The link format: `http://localhost:8080/auth/verify?token=<TOKEN>`

**Direct verification via curl:**
```bash
curl "http://localhost:8080/auth/verify?token=<YOUR_TOKEN>"
```

### Step 3: Login

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "securepass123"
  }'
```

Response:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "user_id": 1,
  "email": "john@example.com"
}
```

### Step 4: Use the Access Token

Include the token in all subsequent requests:

```bash
curl http://localhost:8081/products/ \
  -H "Authorization: Bearer <YOUR_ACCESS_TOKEN>"
```

---

## API Reference

### User Service (Port 8080)

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/auth/register` | Register new user | No |
| POST | `/auth/login` | Login and get access token | No |
| GET | `/auth/verify?token=` | Verify email address | No |

#### Register Request Body
```json
{
  "name": "string (required, max 255)",
  "surname": "string (required, max 255)",
  "email": "string (required, valid email)",
  "password": "string (required, min 8 chars)"
}
```

#### Login Request Body
```json
{
  "email": "string (required, valid email)",
  "password": "string (required, min 8 chars)"
}
```

### Product Service (Port 8081)

All endpoints require `Authorization: Bearer <token>` header.

#### Products

| Method | Endpoint | Description | Role Required |
|--------|----------|-------------|---------------|
| GET | `/products/` | List all products | Any authenticated |
| GET | `/products/:id` | Get product by ID | Any authenticated |
| POST | `/products/` | Create new product | Seller + Active |
| PUT | `/products/:id` | Update product | Seller + Active |
| DELETE | `/products/:id` | Delete product | Seller + Active |

**Query Parameters for List Products:**
- `q` — Search query (name/description)
- `category` — Filter by category slug
- `min_price` — Minimum price filter
- `max_price` — Maximum price filter
- `sort` — Sort order
- `limit` — Results per page (1-100, default 20)
- `offset` — Pagination offset

**Create/Update Product Request Body:**
```json
{
  "name": "string (required, max 255)",
  "description": "string (max 1000)",
  "price": "number (required, > 0)",
  "category_ids": [1, 2]
}
```

#### Shopping Cart

| Method | Endpoint | Description | Role Required |
|--------|----------|-------------|---------------|
| GET | `/products/cart/` | Get user's cart | Active |
| POST | `/products/cart/` | Add item to cart | Active |
| PUT | `/products/cart/:product_id` | Update item quantity | Active |
| DELETE | `/products/cart/:product_id` | Remove item from cart | Active |

**Add to Cart Request Body:**
```json
{
  "product_id": 1,
  "quantity": 2
}
```

**Update Cart Item Request Body:**
```json
{
  "quantity": 3
}
```

---

## User Roles

| Role | Permissions |
|------|-------------|
| `buyer` | Browse products, manage cart (default role) |
| `seller` | All buyer permissions + create/update/delete products |

---

## Troubleshooting

### MailHog Not Receiving Emails

If verification emails are not appearing in MailHog:

1. **Try resending** —  create a new account several times

2. **Restart notifications-service:**
   ```bash
   docker-compose restart notification-service
   ```

3. **Check notifications-service logs:**
   ```bash
   docker-compose logs -f notification-service
   ```

4. **Verify Kafka is healthy:**
   ```bash
   docker-compose logs kafka | tail -20
   ```

### Services Not Starting

1. **Check if ports are available:**
   - 5433 (PostgreSQL)
   - 8080 (user-service)
   - 8081 (product-service)
   - 8025, 1025 (MailHog)
   - 9092 (Kafka)

2. **Rebuild from scratch:**
   ```bash
   docker-compose down -v
   docker-compose up --build
   ```

### Database Connection Issues

If services can't connect to PostgreSQL, ensure the database is healthy:
```bash
docker-compose ps
```

All services should show `healthy` status.

---

## Project Architecture
<img width="2502" height="1083" alt="image" src="https://github.com/user-attachments/assets/681e528e-7005-41cf-9d48-824146ff42e0" />


---

## Tech Stack

- **Language:** Go 1.21+
- **Framework:** Gin
- **ORM:** GORM
- **Database:** PostgreSQL 16
- **Message Broker:** Apache Kafka 3.7
- **Email Testing:** MailHog
- **Authentication:** JWT
- **Containerization:** Docker

---

## License

MIT
