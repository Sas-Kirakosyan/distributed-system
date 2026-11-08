# ---- build stage ----
FROM golang:1.27-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/order-service .

# ---- runtime stage ----
FROM gcr.io/distroless/static-debian13:nonroot

COPY --from=builder /out/order-service /order-service
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/order-service"]