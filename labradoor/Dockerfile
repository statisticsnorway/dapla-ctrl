ARG GO_VERSION=""
FROM golang:${GO_VERSION}alpine AS builder
WORKDIR /src
COPY go.* /src/
RUN go mod download
COPY . /src
RUN go build -o bin/labradoor ./cmd/dapla-ctrl-labradoor

FROM gcr.io/distroless/base
WORKDIR /app
COPY --from=builder /src/bin/labradoor /app/labradoor
ENTRYPOINT ["/app/labradoor"]


