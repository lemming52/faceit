FROM google/cloud-sdk:alpine
COPY ./main ./main

RUN apk add --no-cache curl ca-certificates && update-ca-certificates

EXPOSE 3000

CMD ["./main"]