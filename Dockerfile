# Dockerfile

FROM golang:1.23.0

RUN apt-get update && apt-get install -y gcc

WORKDIR /app

COPY . .

RUN go build -o out

CMD ["./out"]