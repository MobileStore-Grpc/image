FROM golang:1.15.8-alpine3.13 as build-env
LABEL maintainer="Arpeet Gupta <arpeet.gupta96@gmail.com>"

RUN mkdir /image
WORKDIR /image
COPY go.mod . 
COPY go.sum .

RUN go mod download
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build  -a -installsuffix cgo -o /go/bin/image cmd/server/*.go

FROM alpine:3.13.1
#RUN apk add --no-cache bash && \
#    apk add --update --no-cache ca-certificates git
COPY --from=build-env /go/bin/image /go/bin/image
RUN mkdir /img
ENTRYPOINT ["/go/bin/image"]
CMD ["--port", "8020"]
# Add `CMD` and Expose port for REST Endpoint 
EXPOSE 8020
