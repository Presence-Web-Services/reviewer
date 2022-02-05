FROM golang:alpine as build-env

RUN mkdir /reviewer
WORKDIR /reviewer
COPY go.mod .
COPY go.sum .

RUN go mod download
COPY . .

RUN go build -o /go/bin/reviewer

FROM alpine
COPY --from=build-env /go/bin/reviewer /go/bin/reviewer
WORKDIR /go/bin
COPY .env ./
EXPOSE 80
ENTRYPOINT ["./reviewer"]
