# Use a imagem base do Go
FROM golang:1.21

# Configure o diretório de trabalho dentro do contêiner
WORKDIR /app

# Copie o arquivo go.mod e go.sum para o contêiner
COPY ./go.mod ./
COPY ./go.sum ./

# Baixe as dependências do módulo Go
RUN go mod download

# Copie os arquivos do aplicativo para o contêiner
COPY ./cmd/go-to-do-list/*.go ./

# Compilar o aplicativo Go
RUN go build -o to-do-list

# Inicializar o aplicativo quando o contêiner iniciar
CMD ["./to-do-list"]
