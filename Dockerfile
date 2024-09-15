# DEPENDENCIES
FROM golang:1.23.0-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

# SOURCE CODE
COPY . ./
RUN go build -o bin/app cmd/main.go

# FINAL STAGE
FROM alpine AS final
COPY --from=builder /app/config /config
COPY --from=builder /app/migrations /migrations
COPY --from=builder /app/bin/app /app
COPY --from=builder /app/.env /.env
RUN mkdir logs
EXPOSE 8080
CMD ["/app"]