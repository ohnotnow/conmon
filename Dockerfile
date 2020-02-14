FROM golang:latest as builder

LABEL maintainer="billybofh@gmail.com"

WORKDIR /app

COPY conmon.go go.mod go.sum ./

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o conmon .


FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/conmon .

# Command to run the executable
CMD ["./conmon"]