# Scalable E-Commerce Platform

A production-ready, scalable e-commerce platform built with **Go** using a **microservices architecture**. Each service is independently deployable, containerized with Docker, and orchestrated via Docker Compose.

> **Project URL**: [roadmap.sh/projects/scalable-ecommerce-platform](https://roadmap.sh/projects/scalable-ecommerce-platform)

## 🏗️ Architecture

```
                    ┌─────────────────┐
                    │   Nginx Gateway │ :8000
                    └────────┬────────┘
            ┌────────┬───────┼───────┬────────┐
            ▼        ▼       ▼       ▼        ▼
       ┌────────┐┌────────┐┌─────┐┌───────┐┌────────┐
       │  User  ││Product ││Cart ││ Order ││Payment │
       │:8081   ││:8082   ││:8083││:8084  ││:8085   │
       └───┬────┘└───┬────┘└──┬──┘└───┬───┘└───┬────┘
           │         │        │       │         │
           ▼         ▼        ▼       │         ▼
       PostgreSQL PostgreSQL Redis    │     PostgreSQL
       (users)    (products)          │     (payments)
                                      ▼
                                  PostgreSQL  ──► RabbitMQ ──► Notification
                                  (orders)                     Service
```

## 🛠️ Tech Stack

| Component        | Technology                |
|------------------|---------------------------|
| Language         | Go 1.22+                  |
| Web Framework    | Gin Gonic                 |
| Database         | PostgreSQL 16             |
| Cache            | Redis 7                   |
| Message Broker   | RabbitMQ 3                |
| API Gateway      | Nginx                     |
| Auth             | JWT (HS256)               |
| ORM              | GORM                      |
| Containerization | Docker & Docker Compose   |

## 📦 Microservices

### 1. User Service (`:8081`)
Handles user registration, authentication, and profile management.

| Method | Endpoint                  | Auth | Description          |
|--------|---------------------------|------|----------------------|
| POST   | `/api/v1/users/register`  | ❌   | Register new user    |
| POST   | `/api/v1/users/login`     | ❌   | Login & get JWT      |
| GET    | `/api/v1/users/me`        | ✅   | Get user profile     |

### 2. Product Service (`:8082`)
Manages product catalog with CRUD and inventory operations.

| Method | Endpoint                        | Auth | Description          |
|--------|---------------------------------|------|----------------------|
| GET    | `/api/v1/products`              | ❌   | List products        |
| GET    | `/api/v1/products/:id`          | ❌   | Get product details  |
| POST   | `/api/v1/products`              | ✅   | Create product       |
| PUT    | `/api/v1/products/:id/stock`    | ✅   | Update stock         |

### 3. Cart Service (`:8083`)
Redis-backed shopping cart management.

| Method | Endpoint                        | Auth | Description          |
|--------|---------------------------------|------|----------------------|
| GET    | `/api/v1/cart`                  | ✅   | Get cart items       |
| POST   | `/api/v1/cart`                  | ✅   | Add item to cart     |
| PUT    | `/api/v1/cart/:productId`       | ✅   | Update quantity      |
| DELETE | `/api/v1/cart/:productId`       | ✅   | Remove item          |
| DELETE | `/api/v1/cart`                  | ✅   | Clear entire cart    |

### 4. Order Service (`:8084`)
Orchestrates the checkout flow across all services.

| Method | Endpoint                        | Auth | Description          |
|--------|---------------------------------|------|----------------------|
| POST   | `/api/v1/orders/checkout`       | ✅   | Place order          |
| GET    | `/api/v1/orders`                | ✅   | List user's orders   |
| GET    | `/api/v1/orders/:id`            | ✅   | Get order details    |

### 5. Payment Service (`:8085`)
Mock payment processing with simulated gateway.

| Method | Endpoint                        | Auth | Description          |
|--------|---------------------------------|------|----------------------|
| POST   | `/api/v1/payments/process`      | ❌   | Process payment      |

### 6. Notification Service (Worker)
Background consumer that listens for order events via RabbitMQ and simulates sending email/SMS notifications.

## 🚀 Getting Started

### Prerequisites
- [Docker](https://docs.docker.com/get-docker/) & [Docker Compose](https://docs.docker.com/compose/install/)

### Run the Platform

```bash
# Clone the repository
git clone https://github.com/<your-username>/scalable-ecommerce-platform.git
cd scalable-ecommerce-platform

# Start all services
docker compose up --build
```

The API Gateway will be available at `http://localhost:8000`.

RabbitMQ Management UI is at `http://localhost:15672` (guest/guest).

### Stop the Platform

```bash
docker compose down
```

To also remove data volumes:
```bash
docker compose down -v
```

## 🧪 Testing the API

### 1. Register a User
```bash
curl -X POST http://localhost:8000/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com","password":"password123"}'
```

### 2. Login
```bash
curl -X POST http://localhost:8000/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"password123"}'
```
Save the `token` from the response.

### 3. Create a Product
```bash
curl -X POST http://localhost:8000/api/v1/products/ \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -d '{"name":"Gaming Laptop","description":"High-end gaming laptop","price":1299.99,"stock":50,"category":"Electronics"}'
```

### 4. Add to Cart
```bash
curl -X POST http://localhost:8000/api/v1/cart/ \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -d '{"product_id":1,"name":"Gaming Laptop","price":1299.99,"quantity":1}'
```

### 5. Checkout
```bash
curl -X POST http://localhost:8000/api/v1/orders/checkout \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### 6. Check Notifications
```bash
docker logs ecommerce-notification-service
```

## 📂 Project Structure

```
.
├── docker-compose.yml          # Container orchestration
├── nginx.conf                  # API Gateway routing
├── init-db.sh                  # PostgreSQL initialization
├── go.work                     # Go workspace
├── user-service/               # User management & auth
│   ├── cmd/api/main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── handler/
│   │   ├── middleware/
│   │   ├── model/
│   │   ├── repository/
│   │   └── service/
│   ├── Dockerfile
│   └── go.mod
├── product-service/            # Product catalog
├── cart-service/               # Shopping cart (Redis)
├── order-service/              # Order orchestration
├── payment-service/            # Payment processing
└── notification-service/       # Event-driven notifications (RabbitMQ)
```

## 🔄 Checkout Flow

```
User ──► Order Service
              │
              ├──► Cart Service (fetch items)
              ├──► Product Service (reduce stock)
              ├──► Payment Service (process payment)
              ├──► Cart Service (clear cart)
              └──► RabbitMQ (publish event)
                       │
                       └──► Notification Service (consume & notify)
```

## 📝 License

This project is licensed under the MIT License.
