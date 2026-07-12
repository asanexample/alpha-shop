# payment — internal Go service. Static distroless-nonroot image (Kyverno-compliant). Context = repo root.
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/payment ./cmd/payment

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/payment /payment
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/payment"]
