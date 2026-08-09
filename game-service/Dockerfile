FROM golang:1.26.4-alpine AS builder

WORKDIR /src

COPY game-service/go.mod game-service/go.sum ./

RUN go mod download

COPY game-service/ .

RUN CGO_ENABLED=0 go build -o /out/server ./cmd/server/

FROM gcr.io/distroless/static:nonroot

COPY --from=builder /out/server /server

ENV LISTEN_ADDR=:8080
ENV MODEL_ID=iter13-wr72

EXPOSE 8080

ENTRYPOINT ["/server"]