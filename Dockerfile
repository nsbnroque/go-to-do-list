FROM golang:1.21

# Set destination for COPY
WORKDIR /app

# Set the value for NEO4J_URI
ENV NEO4J_URI="bolt://localhost:7687"

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code. 
COPY . .

# Specify the working directory where you want to run the build command
WORKDIR /app/cmd/go-to-do-list

# Compilar o aplicativo Go
RUN go build

EXPOSE 8081

# Inicializar o aplicativo quando o contêiner iniciar
CMD ["./go-to-do-list"]
