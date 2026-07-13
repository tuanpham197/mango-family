# Build stage
FROM golang:1.25-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /server .

# Run stage
FROM alpine:3.20
WORKDIR /app
COPY --from=build /server /app/server
# web/dist copy vào image ở bước CI (Go serve SPA — design §3)
ENV PORT=8080
EXPOSE 8080
ENTRYPOINT ["/app/server"]
