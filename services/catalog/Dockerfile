# catalog — internal Go service. Static distroless-nonroot image (Kyverno-compliant: non-root, no shell).
# Build context is the repo root (the whole module); dockerfile path is services/catalog/Dockerfile.
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/catalog ./cmd/catalog

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/catalog /catalog
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/catalog"]
