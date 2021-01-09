FROM golang:1.13-alpine AS build

RUN mkdir /faceit
WORKDIR /faceit

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go build -o faceit .

FROM alpine:3.9
RUN apk add ca-certificates

COPY --from=build /faceit/faceit /faceit
RUN mkdir docs
COPY --from=build /faceit/docs/index.html /docs/index.html

EXPOSE 3000
EXPOSE 4566
CMD ["/faceit"]

