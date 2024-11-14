
FROM golang:1.23.0


WORKDIR /app


COPY go.mod ./
COPY go.sum ./


RUN apt-get update && apt-get install -y gcc
RUN go mod download


COPY . .

# Compilar o binário da aplicação
RUN go build -o main

# Expor a porta em que sua aplicação roda
EXPOSE 8080

# Executar o binário
CMD ["./main"]
