# Stage build
FROM golang:alpine3.11 as builder

# Copy the contents from host to the container
COPY . /unpackker-cli

# Switching working directory to application folder
WORKDIR /unpackker-cli

# Building the application using `go build`
ENV GO111MODULE=on
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o unpackker

# Second stage
FROM alpine:3.11

WORKDIR /root/

# Copying artifact from builder to end container
COPY --from=builder /unpackker-cli/unpackker unpackker

# Starting 
ENTRYPOINT [ "./unpackker" ]