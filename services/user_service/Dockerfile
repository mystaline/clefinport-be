# Build stage
FROM golang:1.24.6 AS builder
WORKDIR /app

COPY go.work go.work.sum ./
COPY pkg/go.mod ./pkg/
COPY services/user_service/go.mod ./services/user_service/
COPY services/wallet_service/go.mod ./services/wallet_service/
COPY services/log_service/go.mod ./services/log_service/

RUN go work sync && go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /bin/user-service ./services/user_service/main.go

# Final stage
FROM scratch
WORKDIR /bin
COPY --from=builder /bin/user-service /bin/
EXPOSE 8080
ENTRYPOINT ["/bin/user-service"]
