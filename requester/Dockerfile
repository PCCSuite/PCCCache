FROM golang:alpine AS build
ENV CGO_ENABLED 0
ADD . /go/src/requester/
WORKDIR /go/src/requester
RUN go build .


FROM alpine

COPY --from=build /go/src/requester/requester /usr/bin

RUN apk --no-cache --no-progress upgrade && \
    apk --no-cache --no-progress add ca-certificates

ENTRYPOINT [ "requester" ]