FROM golang:1.21.1

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o rtsl_exporter

EXPOSE 8080
CMD ["./rtsl_exporter"]
