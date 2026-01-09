FROM golang:1.21-alpine AS build
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -o /ticketd .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates sqlite
WORKDIR /app
COPY --from=build /ticketd /app/ticketd
EXPOSE 8080
ENV TICKETD_PORT=8080
CMD ["/app/ticketd"]
