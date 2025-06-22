# Build stage
FROM golang:1.23 AS builder

WORKDIR /builder

COPY go.mod go.sum ./
RUN go mod download

# only copy .go files

COPY . .

# build server
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# build migration
RUN CGO_ENABLED=0 GOOS=linux go build -o migration-tool ./cmd/migration

FROM alpine:latest

WORKDIR /app
ENV TZ=Asia/Makassar

COPY --from=builder /builder/server .
COPY --from=builder /builder/migration-tool .

COPY scripts/run.sh .

RUN chmod +x run.sh

EXPOSE 3000

CMD ["./run.sh"]