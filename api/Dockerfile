FROM golang:1.24.0-alpine3.20

WORKDIR app/api/

COPY ./api .
COPY ./.env ../.env
RUN go mod download && go mod verify

COPY ../messages/ ../messages

EXPOSE 3001
RUN go build main.go
CMD ["./main"]
