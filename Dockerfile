FROM golang:alpine as build-env

RUN mkdir /inquirer
WORKDIR /inquirer
COPY go.mod .
COPY go.sum .

RUN go mod download
COPY . .

RUN go build -o /go/bin/inquirer

FROM alpine
COPY --from=build-env /go/bin/inquirer /go/bin/inquirer
WORKDIR /go/bin
COPY .env ./
EXPOSE 80
ENTRYPOINT ["./inquirer"]
