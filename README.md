# Financial Record System

A clean architecture financial record system built with Go that integrates with WhatsApp (Fontee) and Google Sheets.

## Architecture

This project follows clean architecture principles with clear separation of concerns:

```
financial-record/
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── domain/
│   │   ├── entities/    # Business entities (Transaction, User)
│   │   └── repository/  # Repository interfaces
│   ├── repository/      # Repository implementations (Google Sheets)
│   ├── transport/
│   │   └── http/        # HTTP handlers and routers
│   └── usecase/         # Business logic (transaction processing)
```

## Features

- **WhatsApp Integration**: Receive transaction commands via Fontee webhook
- **Google Sheets Storage**: All transactions are recorded in Google Sheets
- **Balance Tracking**: Real-time balance calculation
- **Transaction History**: View all transactions for a phone number
- **Clean Architecture**: Separated layers for maintainability and testability

## Flow

1. User sends `+500` or `-200` via WhatsApp
2. Fontee calls the webhook API
3. Golang processes the transaction
4. Transaction is recorded in Google Sheets
5. Response is sent back to Fontee

## Prerequisites

- Go 1.21 or higher
- Google Cloud Project with Google Sheets API enabled
- Google Service Account credentials
- Google Sheet with appropriate sharing permissions

## Setup

### 1. Google Cloud Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing one
3. Enable Google Sheets API
4. Create a Service Account:
   - Go to IAM & Admin > Service Accounts
   - Click "Create Service Account"
   - Grant appropriate roles
   - Create and download JSON key

### 2. Google Sheets Setup

1. Create a new Google Sheet
2. Note the Spreadsheet ID from the URL (e.g., `1BxiMvs0XRA5nFMdKvBdBZjGMUUqpt35`)
3. Share the sheet with your Service Account email:
   - Click "Share"
   - Enter Service Account email
   - Grant "Editor" permission

### 3. Environment Configuration

Create a `.env` file in the project root:

```bash
SERVER_PORT=8080
GOOGLE_CREDENTIALS='{"type":"service_account",...}'  # Paste your JSON credentials
SPREADSHEET_ID=your_spreadsheet_id
SHEET_NAME=Transactions
```

### 4. Install Dependencies

```bash
cd financial-record
go mod tidy
```

### 5. Run the Server

```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:8080`

## API Endpoints

### Webhook (Fontee Integration)
```
POST /api/webhook
Content-Type: application/json

{
  "phone": "+1234567890",
  "message": "+500"
}
```

Response:
```json
{
  "success": true,
  "message": "Transaction successful. New balance: 500.00",
  "phone": "+1234567890",
  "action": "add",
  "amount": 500,
  "balance": 500,
  "timestamp": "2024-06-17 11:00:00"
}
```

### Get Balance
```
GET /api/balance?phone=+1234567890
```

Response:
```json
{
  "success": true,
  "phone": "+1234567890",
  "balance": 500
}
```

### Get Transaction History
```
GET /api/history?phone=+1234567890
```

Response:
```json
{
  "success": true,
  "phone": "+1234567890",
  "transactions": [...]
}
```

### Health Check
```
GET /health
```

Response: `OK`

## Fontee Configuration

Configure Fontee to call your webhook:

- **Webhook URL**: `https://your-domain.com/api/webhook`
- **Method**: POST
- **Headers**: 
  - `Content-Type: application/json`

## Message Format

Users can send transactions via WhatsApp using these formats:

- `+500` - Add 500 to balance
- `-200` - Subtract 200 from balance

## Google Sheets Structure

The system automatically creates a sheet with the following columns:

| Date | Phone | Action | Amount | Balance | Notes |
|------|-------|--------|--------|---------|-------|
| 2024-06-17 11:00:00 | +123... | add | 500 | 500 | - |

## Security Considerations

- **API Authentication**: Add authentication middleware for production
- **Rate Limiting**: Implement rate limiting to prevent abuse
- **Input Validation**: All inputs are validated
- **HTTPS**: Use HTTPS in production
- **Environment Variables**: Never commit credentials to version control

## Development

### Project Structure

```
financial-record/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── domain/
│   │   ├── entities/
│   │   │   ├── transaction.go   # Transaction entity
│   │   │   └── user.go          # User entity
│   │   └── repository/
│   │       └── transaction_repository.go  # Repository interface
│   ├── repository/
│   │   └── google_sheets_repository.go    # Google Sheets implementation
│   ├── transport/
│   │   └── http/
│   │       ├── handler.go       # HTTP handlers
│   │       └── router.go        # HTTP router
│   └── usecase/
│       └── transaction_usecase.go  # Business logic
├── go.mod
├── go.sum
└── README.md
```

### Testing

```bash
go test ./...
```

## Deployment

### Docker

Create a `Dockerfile`:

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o server cmd/server/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/server .
CMD ["./server"]
```

Build and run:

```bash
docker build -t financial-record .
docker run -p 8080:8080 --env-file .env financial-record
```

## License

MIT License
