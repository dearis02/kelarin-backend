# Build stage
FROM golang:1.23 AS builder

WORKDIR /builder

COPY go.mod go.sum ./
RUN go mod download

# only copy .go files

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

FROM alpine:latest

WORKDIR /app
ENV TZ=Asia/Makassar

COPY --from=builder /builder/server .

EXPOSE 3000

CMD ["./server"]