FROM golang:alpine

WORKDIR /app
COPY . /app

RUN go build -o main .

ENV GIN_MODE=release

EXPOSE 8000

ENTRYPOINT ["./main"]