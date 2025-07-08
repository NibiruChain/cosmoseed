FROM golang:1.24-alpine AS builder

WORKDIR /workspace
COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 go build -o cosmoseed ./cmd/cosmoseed

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/cosmoseed .
USER nonroot:nonroot
ENTRYPOINT ["/cosmoseed"]