# Chat Service

Go service that exposes:

- HTTP endpoint `/chat` (REST)
- gRPC service `ChatService.ChatStream` (server streaming)
- MySQL persistence for chats/messages

## Requirements

- Go 1.24+
- Docker + Docker Compose
- `protoc` + Go plugins (for gRPC codegen)

## Configuration

Create a `.env` in the project root:

```bash
DB_DRIVER=mysql
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=root
DB_PASSWORD=root
DB_NAME=chat_app
WEB_SERVER_PORT=8080
GRPC_SERVER_PORT=50052
INITIAL_CHAT_MESSAGE=You are a helpful assistant.
OPENAI_API_KEY=sk-...
MODEL=gpt-3.5-turbo
MODEL_MAX_TOKENS=4096
TEMPERATURE=0.7
TOP_P=1
N=1
STOP=
MAX_TOKENS=512
AUTH_TOKEN=secrettoken123
```

Notes:

- `MODEL` must be supported by `tiktoken-go` (ex: `gpt-3.5-turbo`, `gpt-4`).
- `AUTH_TOKEN` is required for HTTP requests.
- gRPC currently has no auth.

## Run MySQL

```bash
docker compose up -d
```

## Run Migrations

```bash
make migrate
```

## Run HTTP Server

```bash
go run cmd/chatservice/main.go
```

## HTTP Request (REST)

```bash
POST http://localhost:8080/chat
Content-Type: application/json
Authorization: Bearer secrettoken123

{
  "ChatID": "1",
  "UserID": "1",
  "UserMessage": "Hello, qual e o seu nome?"
}
```

## gRPC Codegen

Install protoc and plugins (Debian-based container example):

```bash
apt-get update && apt-get install -y protobuf-compiler
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
export PATH="$PATH:$(go env GOPATH)/bin"
```

Generate stubs:

```bash
make grpc
```

## gRPC (Postman)

Reflection is enabled in the server. Use:

```bash
dns:localhost:50052
```

Then select `ChatService` -> `ChatStream`.

## Troubleshooting

- If MySQL fails to start with root user error, ensure `docker-compose.yaml` does not set `MYSQL_USER=root`.
- If you get `connection refused`, confirm the container is running: `docker ps`.
- If HTTP returns `401`, check `Authorization: Bearer <AUTH_TOKEN>`.
- If OpenAI returns `401` or `429`, check API key/billing.
