FROM golang:latest as builder

LABEL maintainer="billybofh@gmail.com"

WORKDIR /go/src

COPY conmon.go ./
RUN go get -v github.com/ohnotnow/conmon/...
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o conmon .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/conmon .

# Command to run the executable
CMD ["./conmon"]