FROM golang:1.22.2-alpine3.19

WORKDIR app/api/

COPY ./api .
COPY ./.env ../.env
RUN go mod download && go mod verify

COPY ../messages/ ../messages

EXPOSE 3001
RUN go build main.go
CMD ["./main"]
