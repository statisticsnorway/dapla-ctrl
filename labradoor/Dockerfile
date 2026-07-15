ARG GO_VERSION="1.26.5"
FROM golang:${GO_VERSION} AS builder
WORKDIR /src
COPY go.* /src/
RUN go mod download
COPY . /src
RUN go build -o bin/labradoor ./cmd

FROM gcr.io/distroless/cc
WORKDIR /app
COPY --from=builder /src/bin/labradoor /app/labradoor
ENTRYPOINT ["/app/labradoor"]
