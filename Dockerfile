FROM golang:1.19.1
WORKDIR /go/src/github.com/nikolai5slo/ttlock2mqtt/
COPY ./ ./
RUN CGO_ENABLED=0 go build -o ttlock2mqtt ./main

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app/
COPY --from=0 /go/src/github.com/nikolai5slo/ttlock2mqtt ./
COPY ./templates ./templates
RUN mkdir data

ENV SERVER_ADDRESS=0.0.0.0:8080
ENV STORAGE_FILE=/app/data/storage.json

ENTRYPOINT ["/app/ttlock2mqtt"]