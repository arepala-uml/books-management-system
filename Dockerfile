# Use an official Go runtime as a parent image
FROM golang:1.23-alpine

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod tidy

COPY . .

RUN go build -o main .

EXPOSE 9210

# Run the executable
CMD ["./main"]
