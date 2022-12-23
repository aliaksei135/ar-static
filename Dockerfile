FROM golang:1.19-alpine

RUN apk add build-base

RUN mkdir /app

COPY . /app/
WORKDIR /app
RUN go build . && chmod +x /app/ar-static

ENTRYPOINT [ "/app/ar-static" ]
