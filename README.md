# WhatsApp Financial Recorder Chatbot

A WhatsApp-based financial recording system built with Go that integrates with WhatsApp (Fontee) and Google Sheets. Users can record income and expenses by sending simple WhatsApp messages with automatic categorization and monthly summary reports.

## Architecture

This project follows clean architecture principles with clear separation of concerns:

```
financial-record/
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── domain/
│   │   ├── entities/    # Business entities (Transaction, Category, Report)
│   │   └── repository/  # Repository interfaces
│   ├── repository/      # Repository implementations (Google Sheets)
│   ├── scheduler/       # Cron job scheduler for monthly reports
│   ├── transport/
│   │   └── http/        # HTTP handlers and routers
│   ├── usecase/         # Business logic
│   └── utils/           # Utility functions (number parsing, helpers)
```

## Features

- **WhatsApp Integration**: Receive transaction commands via Fontee webhook
- **Indonesian Number Format**: Support for Indonesian number format (e.g., `3.000.000`)
- **Automatic Categorization**: Smart keyword-based transaction categorization
- **Monthly Sheet Management**: Automatic creation of new sheets each month
- **Balance Carry-over**: Automatic balance transfer between months
- **Monthly Summary Reports**: End-of-month financial reports with category breakdowns
- **Scheduled Reports**: Cron job scheduler for automatic monthly report generation
- **Google Sheets Storage**: All transactions stored in Google Sheets with category support
- **Category Management**: Full CRUD operations for categories and keyword mappings
- **Clean Architecture**: Separated layers for maintainability and testability

## Flow

1. User sends `+3.000.000 gaji bulan Juni` or `-25.000 kopi pagi` via WhatsApp
2. Fontee calls the webhook API
3. System parses Indonesian number format and categorizes transaction automatically
4. Transaction is recorded in Google Sheets with category assignment
5. Response is sent back to Fontee
6. At end of month, system generates and sends summary report

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
DEFAULT_PHONE=+1234567890  # Default phone number for single-user mode and scheduled reports
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
  "sender": "+1234567890",
  "message": "+3.000.000 gaji bulan Juni"
}
```

Response:
```json
{
  "success": true,
  "message": "Success",
  "phone": "+1234567890",
  "action": "add",
  "amount": 3000000,
  "balance": 3000000,
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
  "balance": 3000000
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

### Get All Categories
```
GET /api/categories
```

Response:
```json
{
  "success": true,
  "categories": [...]
}
```

### Create Category
```
POST /api/categories/create
Content-Type: application/json

{
  "name": "Food",
  "description": "Makanan dan minuman",
  "color": "#FF6B6B"
}
```

### Get All Keywords
```
GET /api/keywords
```

Response:
```json
{
  "success": true,
  "keywords": [...]
}
```

### Create Keyword
```
POST /api/keywords/create
Content-Type: application/json

{
  "keyword": "makan",
  "category_id": "Food",
  "priority": 10
}
```

### Generate Monthly Report
```
GET /api/reports/monthly?phone=+1234567890&month=June&year=2024
```

Response:
```json
{
  "success": true,
  "report": {
    "phone": "+1234567890",
    "month": "June",
    "year": 2024,
    "total_income": 5000000,
    "total_expense": 2500000,
    "net_balance": 2500000,
    "category_breakdown": [...]
  }
}
```

### Send Monthly Report via WhatsApp
```
POST /api/reports/send
Content-Type: application/json

{
  "phone": "+1234567890",
  "month": "June",
  "year": 2024
}
```

Response:
```json
{
  "success": true,
  "message": "Report generated successfully",
  "phone": "+1234567890",
  "whatsapp_message": "📊 *LAPORAN BULANAN - June 2024*..."
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

Users can send transactions via WhatsApp using Indonesian number format:

- `+3.000.000 gaji bulan Juni` - Add 3,000,000 with description
- `-25.000 kopi pagi` - Subtract 25,000 with description
- `+500.000 bonus kinerja` - Add 500,000 with description

The system automatically:
- Parses Indonesian number format (dots as thousand separators)
- Categorizes transactions based on keywords in description
- Records transactions in the appropriate monthly sheet
- Maintains running balance

## Google Sheets Structure

The system automatically creates multiple sheets:

### Monthly Transaction Sheets
Each month gets its own sheet with the following columns:

| Date | Phone | Action | Amount | Balance | Category | Notes |
|------|-------|--------|--------|---------|----------|-------|
| 2024-06-17 11:00:00 | +123... | add | 3000000 | 3000000 | Income | gaji bulan Juni |

### Categories Sheet
Contains category definitions:

| ID | Name | Description | Color |
|----|------|-------------|-------|
| 20240617110000 | Food | Makanan dan minuman | #FF6B6B |
| 20240617110001 | Transportation | Transportasi dan perjalanan | #4ECDC4 |

### Keywords Sheet
Contains keyword-to-category mappings:

| ID | Keyword | CategoryID | Priority |
|----|---------|------------|----------|
| 20240617110000 | makan | Food | 10 |
| 20240617110001 | kopi | Food | 9 |
| 20240617110002 | grab | Transportation | 9 |

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
│   │   │   ├── category.go      # Category and keyword entities
│   │   │   └── report.go        # Monthly report entity
│   │   └── repository/
│   │       ├── transaction_repository.go  # Transaction repository interface
│   │       ├── category_repository.go     # Category repository interface
│   │       └── report_repository.go       # Report repository interface
│   ├── repository/
│   │   ├── google_sheets_repository.go          # Google Sheets transaction implementation
│   │   └── google_sheets_category_repository.go # Google Sheets category implementation
│   ├── scheduler/
│   │   └── monthly_report_scheduler.go          # Cron job for monthly reports
│   ├── transport/
│   │   └── http/
│   │       ├── handler.go       # HTTP handlers
│   │       └── router.go        # HTTP router
│   ├── usecase/
│   │   ├── transaction_usecase.go  # Transaction business logic
│   │   ├── category_usecase.go     # Category business logic
│   │   └── report_usecase.go       # Report business logic
│   └── utils/
│       ├── number_parser.go       # Indonesian number format parsing
│       ├── number_parser_test.go  # Number parser tests
│       └── helper.go              # Helper functions
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
