FROM golang:1.21-alpine AS builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY . .
RUN go mod tidy && CGO_ENABLED=1 go build -o server .

FROM alpine:latest
RUN apk add --no-cache libc6-compat
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/series.db .
COPY --from=builder /app/swagger.yaml .
EXPOSE 8080
CMD ["./server"]