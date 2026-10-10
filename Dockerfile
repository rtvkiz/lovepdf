# syntax=docker/dockerfile:1

# ---- Build stage ----
FROM docker.io/library/golang:1.24-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/lovepdf ./cmd/server

# Writable directories for the non-root runtime user
RUN mkdir -p /out/tmp /out/.config

# ---- Runtime stage ----
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=build --chown=nonroot:nonroot /out/tmp ./tmp
COPY --from=build --chown=nonroot:nonroot /out/.config ./.config
COPY --from=build /out/lovepdf ./lovepdf
# Templates and static assets are loaded relative to the working directory
COPY web ./web

# pdfcpu writes its configuration under the user config directory
ENV XDG_CONFIG_HOME=/app/.config

USER nonroot
EXPOSE 8080

ENTRYPOINT ["/app/lovepdf"]
CMD ["-port", ":8080", "-tmp", "/app/tmp"]
