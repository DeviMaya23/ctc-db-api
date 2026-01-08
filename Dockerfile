FROM golang:1.24.4-alpine3.22 AS builder
WORKDIR /app
COPY . .

RUN go mod download
RUN go mod tidy

RUN go build -o main main.go

FROM alpine:3.22
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/config.env .

EXPOSE 9080

CMD [ "/app/main" ]