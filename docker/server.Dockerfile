# Build stage
FROM golang:1.23 AS builder

WORKDIR /builder

COPY . .

RUN ls

RUN go mod download

# build server
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# build migration
RUN CGO_ENABLED=0 GOOS=linux go build -o migration-tool ./cmd/migration

# build init area
RUN CGO_ENABLED=0 GOOS=linux go build -o init-area ./cmd/init_area

FROM alpine:latest

WORKDIR /app
ENV TZ=Asia/Makassar

COPY --from=builder /builder/server .
COPY --from=builder /builder/migration-tool .
COPY --from=builder /builder/init-area .

COPY --from=builder /builder/scripts/run.sh .

RUN apk --no-cache add ca-certificates tzdata bash

RUN chmod +x run.sh

EXPOSE 3000

CMD ["./run.sh"]