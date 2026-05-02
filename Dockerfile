# ---- Build Stage ----

# this specifies the base image
# as tells what we will call it in the rest of the file
# also it is beneficial for multi stage builds as if we don't name our stages we have to use integers
# which depends on the number of from statements starting from integer 0
# this is a issue because the order might change and file can break if the necessary changes are not made
FROM golang:1.24-alpine AS builder 

# working directory inside the container
WORKDIR /app

# this below instructions tells the builder to copy 
# the go.mod and go.sum files to the working directory
COPY go.mod go.sum ./

#run this specific command [builder will run this]
RUN go mod download

# copy <source> <destination> - 
# here we copy everything from current directory on host to working directory in the image
# before this we need to set the work directory as if we don't
# destination will be /* wjicj means we are polluting the root directory.
# if we need to exclude anything add to dockerignore file
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -o mini-redis ./cmd/server

# ---- Runtime Stage ----
FROM gcr.io/distroless/base-debian12

WORKDIR /app
# By default, the COPY instruction copies files from the build context. 
# The COPY --from flag lets you copy files from an image, a build stage, 
# or a named context instead.
# here we are copying from our build stage
COPY --from=builder /app/mini-redis .

#this is the port we will export
EXPOSE 6379
ENTRYPOINT ["/app/mini-redis"]
