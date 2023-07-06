FROM golang:1.19
WORKDIR /app
COPY . .
WORKDIR /app
ENV GOSUMDB=off
RUN CGO_ENABLED=0 go build -mod=vendor -o aws-connector cmd/main.go 

# cannot use scratch becaue of the ca-certificates & hosntame -i command used by the service
FROM alpine:3.14 
RUN apk add -U --no-cache ca-certificates
COPY --from=0 /app/aws-connector /
CMD ["/aws-connector"]