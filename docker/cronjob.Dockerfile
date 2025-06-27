# Build stage
FROM golang:1.23 AS builder

WORKDIR /builder

COPY .. .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o cronjob ./cmd/cronjob

FROM alpine:latest
WORKDIR /app
ENV TZ=Asia/Makassar

COPY --from=builder /builder/cronjob .

RUN apk --no-cache add ca-certificates tzdata bash

ARG TASK_NAME

CMD [ "./cronjob" ]