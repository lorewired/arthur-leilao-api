# Use a imagem oficial do Go como base para a construção do projeto
FROM golang:1.21.2 AS build

# Definindo o diretório de trabalho dentro do container
WORKDIR /app

# Copiando go.mod e go.sum
COPY go.mod go.sum ./

# Baixando as dependências Go
RUN go mod download

# Copiando o restante dos arquivos do projeto
COPY . .

# Construindo o binário Go
RUN go build -o out ./cmd

# Usando uma imagem mais leve para o runtime
FROM ubuntu:jammy

# Definindo o diretório de trabalho
WORKDIR /app

# Copiando o binário construído da etapa anterior
COPY --from=build /app/out /app/out

# Definindo o comando de entrada
CMD ["/app/out"]
