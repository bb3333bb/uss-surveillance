FROM golang:1.26-bookworm AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/gateway ./cmd/gateway

FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=builder /out/gateway ./gateway
EXPOSE 8080
ENTRYPOINT ["./gateway"]
