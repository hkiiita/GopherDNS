FROM golang:1.23

WORKDIR /app

COPY . .

RUN go build -o dnsgopher .
EXPOSE 53/udp
CMD ["./dnsgopher"]