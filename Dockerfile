FROM python:alpine

RUN set -x \
    && apk add --no-cache supercronic shadow

COPY src/ .
COPY requeriments.txt .
COPY crontab .

RUN pip install -r requeriments.txt

CMD ["supercronic", "crontab"]