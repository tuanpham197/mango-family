# Build stage
FROM golang:1.25-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Chỉ build API (web deploy riêng trên Vercel). Static binary, không CGO.
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /server .

# Run stage
FROM alpine:3.20
WORKDIR /app
# ca-certificates: cần để xác thực TLS khi kết nối Supabase (sslmode=require).
RUN apk add --no-cache ca-certificates
COPY --from=build /server /app/server
ENV PORT=8080
EXPOSE 8080
ENTRYPOINT ["/app/server"]
