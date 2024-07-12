# Etapa de base, usando uma imagem Ubuntu como base
FROM ubuntu:jammy AS base

# Definindo o diretório de trabalho
WORKDIR /app

# Etapa de configuração do Nix
FROM base AS nix-config

# Copiando o arquivo de configuração do Nix
COPY .nixpacks/nixpkgs-e89cf1c932006531f454de7d652163a9a5c86668.nix .nixpacks/nixpkgs-e89cf1c932006531f454de7d652163a9a5c86668.nix

# Instalando as dependências do Nix e limpando o cache
RUN nix-env -if .nixpacks/nixpkgs-e89cf1c932006531f454de7d652163a9a5c86668.nix && nix-collect-garbage -d

# Etapa de construção
FROM nix-config AS build

# Copiando todos os arquivos do projeto para o diretório de trabalho
COPY . /app

# Definindo o cache para o Go
RUN --mount=type=cache,id=s/d6e68f11-4a86-452e-bfd9-2e7ad760fc72-/root/cache/go-build,target=/root/.cache/go-build go mod download

# Construindo o binário Go
RUN --mount=type=cache,id=s/d6e68f11-4a86-452e-bfd9-2e7ad760fc72-/root/cache/go-build,target=/root/.cache/go-build go build -o out ./...

# Etapa final
FROM ubuntu:jammy

# Definindo o diretório de trabalho
WORKDIR /app

# Copiando o binário construído da etapa anterior
COPY --from=build /app/out /app/out

# Definindo o comando de entrada
CMD ["/app/out"]
