FROM golang:1.26.4 AS build-stage

WORKDIR /app

COPY go.mod ./
COPY ./cmd/ ./cmd/
COPY ./internal/ ./internal/

RUN CGO_ENABLED=0 GOOS=linux go build -o /triv_lb ./cmd/triv_lb/main.go

# Run tests
FROM build-stage AS run-test-stage
RUN go test -v ./...

FROM gcr.io/distroless/base-debian11 AS build-release-stage

COPY --from=build-stage /triv_lb /triv_lb
EXPOSE 8080

ENTRYPOINT ["/triv_lb", "--port", "8080", "--config", "/config/config.json"]