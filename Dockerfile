# Start from golang base image
FROM golang:alpine as builder

# ENV GO111MODULE=on

RUN apk update && apk add --no-cache git && apk add build-base && apk add libmediainfo-dev

# Set the current working directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and the go.sum files are not changed
RUN go mod download && go mod verify

# Copy the source from the current directory to the working Directory inside the container
COPY . .

# Build the Go app
#RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo cmd/main/main.go
RUN GOOS=linux go build cmd/main/main.go

# Start a new stage from scratch
FROM alpine:latest

RUN apk --no-cache add ca-certificates
RUN apk add libmediainfo

WORKDIR /root/

#ADD https://github.com/ufoscout/docker-compose-wait/releases/download/2.9.0/wait wait
#RUN chmod +x wait

# Copy the Pre-built binary file from the previous stage. Observe we also copied the .env file
COPY --from=builder /app/main .
COPY --from=builder /app/.env .

#Command to run the executable
#CMD ["./wait"]
CMD ["./main"]