# DivvyDoo Backend

A robust backend service for DivvyDoo - a divvydoo/backend-like expense sharing application built with Go.

## 📋 Table of Contents

- [Overview](#overview)
- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
- [API Documentation](#api-documentation)
- [Architecture](#architecture)
- [Development](#development)
- [Environment Variables](#environment-variables)

## 🎯 Overview

DivvyDoo Backend is a RESTful API service that powers expense sharing and settlement between users and groups. It provides features like:

- User authentication and authorization
- Group management
- Expense tracking and splitting
- Balance calculation and settlement
- Real-time balance updates via worker

## 🛠 Tech Stack

- **Language**: Go 1.21+
- **Database**: MongoDB
- **Authentication**: JWT
- **Architecture**: Clean Architecture (Controllers → Services → Repositories)

## 📁 Project Structure

```
backend/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── controllers/             # HTTP request handlers
│   │   ├── balance.go
│   │   ├── expense.go
│   │   ├── group.go
│   │   ├── settlement.go
│   │   └── user.go
│   ├── middleware/              # HTTP middleware
│   │   └── auth.go             # JWT authentication middleware
│   ├── models/                  # Domain models
│   │   ├── balance.go
│   │   ├── expense.go
│   │   ├── group.go
│   │   ├── settlement.go
│   │   └── user.go
│   ├── repositories/            # Data access layer
│   │   ├── balance.go
│   │   ├── expense.go
│   │   ├── group.go
│   │   ├── settlement.go
│   │   └── user.go
│   ├── services/                # Business logic layer
│   │   ├── balance.go
│   │   ├── expense.go
│   │   ├── group.go
│   │   ├── settlement.go
│   │   └── user.go
│   ├── utils/                   # Utility functions
│   │   ├── errors.go
│   │   └── responses.go
│   └── worker/                  # Background workers
│       └── balance_worker.go
├── pkg/
│   └── auth/
│       └── jwt.go              # JWT token management
├── go.mod                       # Go module definition
└── README.md                    # This file
```

## 🚀 Getting Started

### Prerequisites

- Go 1.21 or higher
- MongoDB 6.0 or higher
- Git

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd DivvyDoo/backend
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up environment variables**
   
   Create a `.env` file in the backend directory:
   ```env
   # Server
   PORT=8080
   ENV=development

   # Database
   MONGODB_URI=mongodb://localhost:27017
   MONGODB_DATABASE=divvydoo

   # JWT
   JWT_SECRET=your-secret-key-here
   JWT_EXPIRY=24h

   # CORS
   ALLOWED_ORIGINS=http://localhost:3000
   ```

4. **Initialize MongoDB**
   
   Run the initialization script:
   ```bash
   mongosh < ../scripts/mongo-init.js
   ```

5. **Run the application**
   ```bash
   go run cmd/api/main.go
   ```

   The server will start on `http://localhost:8080`

### Running Tests

```bash
go test ./...
```

### Building for Production

```bash
go build -o bin/api cmd/api/main.go
./bin/api
```

## 📚 API Documentation

### Base URL
```
http://localhost:8080/v1
```

### Authentication

All authenticated endpoints require a JWT token in the Authorization header:
```
Authorization: Bearer <token>
```

### Endpoints

#### Authentication & Users
**Public:**
- `POST /v1/login` - User login
- `POST /v1/users` - Create a new user (register)

**Authenticated:**
- `GET /v1/users/:id` - Get user details
- `PUT /v1/users/:id` - Update user

#### Groups
**All endpoints require authentication**
- `POST /v1/groups` - Create a new group
- `GET /v1/groups/:id` - Get group details
- `POST /v1/groups/:id/members` - Add member to group

#### Expenses
**All endpoints require authentication**
- `POST /v1/expenses` - Create a new expense
- `GET /v1/expenses/:id` - Get expense details
- `GET /v1/groups/:id/expenses` - List all expenses for a group
- `GET /v1/users/:id/expenses` - List all expenses for a user

#### Balances
**All endpoints require authentication**
- `GET /v1/users/:id/balances` - Get all balances for a user
- `GET /v1/groups/:id/balances` - Get all balances for a group

#### Settlements
**All endpoints require authentication**
- `POST /v1/settlements` - Create a new settlement
- `GET /v1/settlements/:id` - Get settlement details

## 🏗 Architecture

This project follows Clean Architecture principles with clear separation of concerns:

### Layers

1. **Controllers Layer** (`internal/controllers/`)
   - Handle HTTP requests/responses
   - Validate input
   - Call appropriate services
   - Return formatted responses

2. **Services Layer** (`internal/services/`)
   - Implement business logic
   - Coordinate between repositories
   - Handle complex operations
   - Trigger background workers

3. **Repository Layer** (`internal/repositories/`)
   - Database operations
   - Data persistence
   - Query building
   - CRUD operations

4. **Models Layer** (`internal/models/`)
   - Domain entities
   - Data structures
   - Business rules

### Data Flow

```
Request → Middleware → Controller → Service → Repository → Database
                                        ↓
                                     Worker
```

## 💻 Development

### Code Style

- Follow Go conventions and idioms
- Use `gofmt` for code formatting
- Use `golint` for linting

### Adding a New Feature

1. Define model in `internal/models/`
2. Create repository in `internal/repositories/`
3. Implement service in `internal/services/`
4. Add controller in `internal/controllers/`
5. Register routes in `cmd/api/main.go`

### Background Workers

The balance worker (`internal/worker/balance_worker.go`) runs asynchronously to:
- Calculate user balances
- Update balance records
- Handle settlement computations

## 🔐 Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `ENV` | Environment (development/production) | `development` |
| `MONGODB_URI` | MongoDB connection string | `mongodb://localhost:27017` |
| `MONGODB_DATABASE` | Database name | `divvydoo` |
| `JWT_SECRET` | Secret key for JWT signing | - |
| `JWT_EXPIRY` | JWT token expiry duration | `24h` |
| `ALLOWED_ORIGINS` | CORS allowed origins | `*` |

## 📝 License

This project is licensed under the MIT License.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📧 Contact

For questions or support, please open an issue in the repository.
