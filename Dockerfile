FROM golang:1.24-alpine AS builder

RUN apk add make git

WORKDIR /workspace
COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    make build

FROM gcr.io/distroless/static:nonroot
WORKDIR /bin
COPY --from=builder /workspace/build/cosmoseed .
USER nonroot:nonroot
ENTRYPOINT ["cosmoseed"]