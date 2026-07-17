FROM golang:1.25 AS builder

LABEL maintainer="Deepak"
LABEL description="Aegis"
LABEL version="1.0.0"

ENV CGO_ENABLED=0
ENV GOOS=linux

WORKDIR /app

COPY apps/server/go.mod apps/server/go.sum ./apps/server/
RUN cd apps/server && go mod download

COPY apps/server/ ./apps/server/
RUN cd apps/server && go generate ./... && go build -v -o ../server ./cmd/server

FROM gcr.io/distroless/static:nonroot

WORKDIR /app
COPY --from=builder /app/apps/server .
COPY --from=builder /app/apps/server/migrations ./migrations

EXPOSE 8080

ENTRYPOINT ["/app/server"]
