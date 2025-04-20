FROM golang:1.24.2-alpine3.21 AS builder

ARG CGO_ENABLED=0
WORKDIR /app

COPY . .
RUN go mod download && go build .

FROM scratch
COPY --from=builder /app/Minerva /Minerva
ENTRYPOINT ["/Minerva"]
