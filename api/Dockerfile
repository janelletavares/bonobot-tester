FROM golang:alpine as builder

WORKDIR /go/src/app
COPY . .
RUN CGO_ENABLED=0 go install

FROM alpine

RUN mkdir /app
COPY --from=builder /go/bin/api /app

RUN mkdir /app/internal
COPY internal/migrations /app/internal/migrations

EXPOSE 80

ENTRYPOINT /app/api serve --http=0.0.0.0:80