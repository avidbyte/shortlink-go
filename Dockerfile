FROM golang:1.22 as builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /shortlink-go cmd/app/main.go

FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=builder /shortlink-go /shortlink-go
COPY config.yaml .
EXPOSE 8080
CMD ["/shortlink-go"]
