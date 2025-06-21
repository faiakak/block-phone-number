# EmagineNET Blocked Phone Number API

This project is a simple web application built in Go for managing blocked phone numbers to prevent check fraud across multiple grocery store locations.

## 📦 Features

- Block phone numbers with reason, store, and metadata
- Check if a phone number is blocked
- List and remove blocked phone numbers
- Manage store locations
- RESTful JSON API
- PostgreSQL storage with auto migrations
- Dockerized deployment 

## 🚀 Getting Started

### Prerequisites

- [Go 1.21+](https://golang.org/dl/)
- [Docker + Docker Compose](https://docs.docker.com/compose/)
- PostgreSQL client (optional)

### Clone the repository

```bash
git clone https://github.com/faiakak/block-phone-number.git
cd block-phone-number
```

---

## 🔧 Local Installation Steps

### 1. Create a `.env` file

```env
DB_HOST=localhost
DB_PORT=5433
DB_USER=postgres
DB_PASSWORD=secret
DB_NAME=emaginenet_blocked_numbers
APP_PORT=8080
ENV=local
```

### 2. Start PostgreSQL (using Docker or your own setup)

```bash
docker-compose up db
```

Or manually create the DB:

```bash
createdb -U postgres -h localhost -p 5433 emaginenet_blocked_numbers
```

### 3. Run the Go app

```bash
go run main.go
```

The server will be available at: `http://localhost:8080`

---

## 🐳 Dockerized Installation Steps

This starts both the Go API and PostgreSQL via Docker Compose.

```bash
docker-compose up --build
```

Ensure `.env` is set as:

```env
DB_HOST=db
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=secret
DB_NAME=emaginenet_blocked_numbers
APP_PORT=8080
ENV=production
```

After successful startup:
- API available at `http://localhost:8080`
- DB available at `localhost:5433`

---

## 📚 API Endpoints

| Method | Endpoint              | Description                  |
|--------|-----------------------|------------------------------|
| GET    | `/api/blocked-phones` | List blocked numbers         |
| POST   | `/api/blocked-phones` | Block a phone number         |
| DELETE | `/api/blocked-phones/{id}` | Unblock a number       |
| POST   | `/api/check-phone`    | Check block status           |
| GET    | `/api/stores`         | List all stores              |
| POST   | `/api/stores`         | Create a new store           |

---

## 🧪 Testing

To run unit tests:

```bash
go test ./...
```

---

## 📁 Project Structure

```
.
├── config/            # DB connection & migrations
├── handlers/          # HTTP route handlers & Test
├── routes/            # Router definitions
├── Dockerfile
├── docker-compose.yml
├── README.md
└── main.go
```


