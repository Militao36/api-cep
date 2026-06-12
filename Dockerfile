FROM golang:1.26-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o api-cep .

FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update \
	&& apt-get install -y --no-install-recommends \
		bash \
		ca-certificates \
		curl \
	&& rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/api-cep .

EXPOSE 8080

CMD ["./api-cep"]
