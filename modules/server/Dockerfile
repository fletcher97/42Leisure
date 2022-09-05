FROM golang:1.19-alpine

COPY . /42Leisure-api
WORKDIR /42Leisure-api
RUN go mod download

CMD go run .
