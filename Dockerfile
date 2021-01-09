FROM golang:1.13-alpine AS build

WORKDIR /dockerdev

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go build -o /faceit .

FROM alpine:3.9
RUN apk add ca-certificates

COPY --from=build /dockerdev/faceit /faceit

EXPOSE 3000
CMD ["/faceit"]

