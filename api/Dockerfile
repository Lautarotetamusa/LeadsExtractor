FROM golang:1.24.1-alpine3.20

RUN apk add --no-cache tzdata

ENV TZ=America/Mexico_City

WORKDIR app/api/

COPY ./api .
COPY ./.env ../.env
RUN go mod download && go mod verify

RUN go build -o daily-reporter cmd/reporter/main.go

EXPOSE 3001
RUN go build main.go

RUN echo "0 8 * * * cd /go/app/api && ./daily-reporter >> /var/log/reporter.log 2>&1" \
    > /var/spool/cron/crontabs/root

RUN chmod +x entrypoint.sh
CMD ["./entrypoint.sh"]
