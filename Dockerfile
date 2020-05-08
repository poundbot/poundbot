FROM golang:1.14.1-alpine as builder

# Move to working directory /build
WORKDIR /build

# Copy the code into the container
COPY . .
RUN go mod download

# Build the application
RUN go build -ldflags="-s -w -X main.version=`cat cmd/poundbot/VERSION` -X main.buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.githash=`git rev-parse --short HEAD`" \
    -o poundbot cmd/poundbot/poundbot.go

FROM alpine:latest

# Move to /dist directory as the place for resulting binary folder
WORKDIR /dist

COPY --from=builder /build/poundbot .
COPY ./templates_sample /dist/templates
COPY ./language /dist/language

# Export necessary ports
EXPOSE 9090 6061

# Command to run when starting the container
CMD [ "/dist/poundbot" ]