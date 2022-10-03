### API Build
FROM golang:alpine AS build
ENV CGO_ENABLED 0
ADD . /go/src/requester/
WORKDIR /go/src/requester
RUN go build .


FROM alpine

## API
COPY --from=build /go/src/requester/requester /usr/bin

### samba
RUN apk --no-cache --no-progress upgrade && \
    apk --no-cache --no-progress add ca-certificates

EXPOSE 8080

## entrypoint
ENTRYPOINT [ "requester" ]