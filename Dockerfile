# syntax=docker/dockerfile:1
#
# Builds the Ladegast binary with the frontend embedded. Runtime
# configuration is entirely via environment variables (see
# schema/config.md / .env.example) - this image has no deployment-specific
# assumptions baked in.

FROM node:22-alpine AS frontend
WORKDIR /src/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ .
# vite.config.ts's outDir ("../internal/webui/dist") resolves relative to
# this WORKDIR, landing at /src/internal/webui/dist - picked up by the
# builder stage below.
RUN npm run build

FROM golang:1.26-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=frontend /src/internal/webui/dist internal/webui/dist
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/ladegast .
# distroless/nonroot runs as uid:gid 65532:65532; pre-create the data dir
# with that ownership so it's writable without running as root.
RUN mkdir -p /data && chown 65532:65532 /data

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=builder /out/ladegast /app/ladegast
COPY --from=builder --chown=65532:65532 /data /data

ENV LADEGAST_DATA_DIR=/data
VOLUME ["/data"]
EXPOSE 8080

USER nonroot:nonroot
ENTRYPOINT ["/app/ladegast"]
