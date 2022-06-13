FROM golang:1.18 AS build
WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
ENV GOSUMDB=off
RUN CGO_ENABLED=0 go build -mod=vendor -o aws-connector cmd/main.go

FROM alpine:3.14
COPY --from=build /usr/src/app/aws-connector /
CMD ["/aws-connector"]