# syntax=docker/dockerfile:1
# Pinned to the build host's native platform rather than the target platform - its
# output is platform-independent static assets, and Bun's JIT is prone to crashing
# under QEMU emulation when the target platform differs from the host's.
FROM --platform=$BUILDPLATFORM oven/bun:1 AS ui-build
WORKDIR /src/web
COPY web/package.json web/bun.lock ./
RUN bun install --frozen-lockfile
COPY web/ ./
RUN bun run build

FROM golang:1.27 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=ui-build /src/internal/web/static ./internal/web/static
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/skygate ./cmd/skygate

FROM gcr.io/distroless/static-debian12:nonroot
ARG BUILD_VERSION=dev
ENV BUILD_VERSION=$BUILD_VERSION
COPY --from=build /out/skygate /skygate
EXPOSE 8080 9090
USER nonroot:nonroot
ENTRYPOINT ["/skygate"]
