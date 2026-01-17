# EnerzyFlow Backend

A robust REST API backend service for EnerzyFlow - an energy drink bottle ordering and management platform. Built with Go (Golang) and PostgreSQL, this backend provides comprehensive order management, user authentication, and company profile management capabilities.

## 🚀 Features

- **OTP-Based Authentication**: Secure email-based OTP authentication system using SendGrid/Resend
- **User Management**: Complete user profile management with role-based access control (RBAC)
- **Order Management**: End-to-end order lifecycle management with status tracking
- **Company Profiles**: Multi-outlet company management with custom labels
- **Payment Processing**: Payment screenshot upload and verification system
- **Invoice Management**: Automated invoice generation and storage
- **Label Management**: Custom bottle label design upload and tracking
- **Order Tracking**: Real-time order status updates and tracking
- **Comment System**: Order-level commenting for communication
- **Cloud Storage**: Cloudinary integration for image/file storage
- **CORS Enabled**: Configured for cross-origin requests

## 🛠️ Tech Stack

- **Language**: Go 1.23.3
- **Web Framework**: Gin (v1.11.0)
- **Database**: PostgreSQL (via pgx/v5)
- **Authentication**: JWT (golang-jwt/v5)
- **Cloud Storage**: Cloudinary
- **Email Service**: SendGrid & Resend
- **Environment**: godotenv for configuration

## 📋 Prerequisites

- Go 1.23.3 or higher
- PostgreSQL database
- Cloudinary account (for image storage)
- SendGrid or Resend account (for email OTP)

## 🔧 Installation

1. **Clone the repository**

   ```bash
   git clone https://github.com/shaukatalidev/enerzyflow_backend.git
   cd enerzyflow_backend
   ```

2. **Install dependencies**

   ```bash
   go mod download
   ```

3. **Set up environment variables**

   Create a `.env` file in the root directory:

   ```env
   # Database
   DB_URL=postgresql://user:password@host:port/database?sslmode=require

   # Cloudinary
   CLOUDINARY_CLOUD_NAME=your_cloud_name
   CLOUDINARY_API_KEY=your_api_key
   CLOUDINARY_API_SECRET=your_api_secret

   # Email Service (SendGrid)
   SENDGRID_API_KEY=your_sendgrid_api_key
   SENDGRID_FROM=your_email@example.com

   # Email Service (Resend - Alternative)
   RESEND_API_KEY=your_resend_api_key
   ```

4. **Run the application**

   ```bash
   go run app.go
   ```

   The server will start on `http://localhost:9080`

## 📁 Project Structure

```
enerzyflow_backend/
├── app.go                      # Application entry point
├── go.mod                      # Go module dependencies
├── go.sum                      # Dependency checksums
├── .env                        # Environment configuration
├── internal/                   # Internal packages
│   ├── auth/                   # Authentication module
│   │   ├── auth_handler.go    # Auth HTTP handlers
│   │   ├── auth_model.go      # Auth data models
│   │   └── auth_service.go    # Auth business logic
│   ├── companies/              # Company management
│   │   ├── company_model.go
│   │   ├── company_repository.go
│   │   └── company_service.go
│   ├── db/                     # Database configuration
│   │   └── db.go
│   ├── orders/                 # Order management
│   │   ├── order_handler.go
│   │   ├── order_model.go
│   │   ├── order_repository.go
│   │   └── order_service.go
│   └── users/                  # User management
│       ├── user_handler.go
│       ├── user_model.go
│       ├── user_repository.go
│       └── user_service.go
├── routes/                     # API route definitions
│   └── router.go
└── utils/                      # Utility functions
    ├── helper.go
    └── jwt.go                  # JWT middleware
```

## 🔌 API Endpoints

### Authentication

```
POST   /auth/send-otp          # Send OTP to email
POST   /auth/verify-otp        # Verify OTP and get JWT token
```

### Users (Protected)

```
POST   /users/profile          # Save/Update user profile
GET    /users/profile          # Get user profile
GET    /users/all              # Get all users (Admin only)
POST   /users/create           # Create user by admin
```

### Orders (Protected)

```
POST   /orders/create                      # Create new order
GET    /orders/get-all                     # Get all orders (for logged-in user)
GET    /orders/:id                         # Get specific order
POST   /orders/:id/payment-screenshot      # Upload payment screenshot
PUT    /orders/:id/status                  # Update order status
PUT    /orders/:id/payment                 # Update payment status (Admin only)
GET    /orders/get-all-orders              # Get all orders (Admin view)
GET    /orders/:id/tracking                # Get order tracking info
POST   /orders/:id/upload-invoice          # Upload invoice (Admin only)
POST   /orders/:id/comment                 # Add comment to order
GET    /orders/:id/comment                 # Get order comments
POST   /orders/:id/label                   # Save label details (Admin only)
GET    /orders/:id/label                   # Get label details
GET    /orders/:id/detail                  # Get detailed order info (Admin only)
```

## 🔐 Authentication & Authorization

The API uses JWT-based authentication with role-based access control:

1. **OTP Flow**: Users receive a one-time password via email
2. **Token Generation**: Upon OTP verification, a JWT token is issued
3. **Protected Routes**: Include the token in the `Authorization` header
   ```
   Authorization: Bearer <your_jwt_token>
   ```
4. **Role-Based Access**: Some endpoints require admin role (e.g., payment updates, invoice uploads)

## 🌐 CORS Configuration

The backend is configured to accept requests from:

- `http://localhost:3000` (Development)
- `https://www.enerzyflow.com` (Production)

## 🗄️ Database

The application uses PostgreSQL with the following main entities:

- Users
- Companies
- Outlets
- Orders
- Labels
- Order Comments
- Order Tracking

## 📦 Key Dependencies

```go
github.com/gin-gonic/gin              // Web framework
github.com/jackc/pgx/v5               // PostgreSQL driver
github.com/golang-jwt/jwt/v5          // JWT authentication
github.com/cloudinary/cloudinary-go   // Cloud storage
github.com/sendgrid/sendgrid-go       // Email service
github.com/resendlabs/resend-go       // Alternative email service
github.com/joho/godotenv              // Environment variables
github.com/google/uuid                // UUID generation
github.com/gin-contrib/cors           // CORS middleware
```

## 🔨 Build & Deploy

### Build for production

```bash
go build -o main app.go
```

### Run the binary

```bash
./main
```

### Docker (Optional)

```dockerfile
FROM golang:1.23.3-alpine
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o main app.go
EXPOSE 9080
CMD ["./main"]
```

## 🧪 Testing

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
```

## 📝 Environment Variables

| Variable                | Description                  | Required |
| ----------------------- | ---------------------------- | -------- |
| `DB_URL`                | PostgreSQL connection string | Yes      |
| `CLOUDINARY_CLOUD_NAME` | Cloudinary cloud name        | Yes      |
| `CLOUDINARY_API_KEY`    | Cloudinary API key           | Yes      |
| `CLOUDINARY_API_SECRET` | Cloudinary API secret        | Yes      |
| `SENDGRID_API_KEY`      | SendGrid API key             | Yes\*    |
| `SENDGRID_FROM`         | SendGrid sender email        | Yes\*    |
| `RESEND_API_KEY`        | Resend API key               | Yes\*    |

\*Either SendGrid or Resend configuration is required

## 🔄 Related Repositories

- **Frontend Repository**: https://github.com/shaukatalidev/enerzyflow_new

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 License

This project is proprietary software. All rights reserved.

## 👥 Authors

- **Shaukat Ali** - [GitHub](https://github.com/shaukatalidev)

## 📧 Contact

For questions or support, please contact: help@enerzyflow.com

## 🙏 Acknowledgments

- Gin Web Framework team
- PostgreSQL community
- All contributors and supporters

---

**Note**: Remember to never commit your `.env` file to version control. Add it to `.gitignore` to keep your credentials secure.
