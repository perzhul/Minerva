FROM golang:1.24.2-alpine3.21

WORKDIR /app

COPY . .

RUN go mod download

COPY *.go ./

RUN go build .

EXPOSE 25565

CMD [ "./Minerva" ]
