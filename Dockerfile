FROM python:alpine

RUN set -x \
    && apk add --no-cache supercronic shadow

WORKDIR app

COPY src/ .
COPY requeriments.txt .
COPY crontab .
#COPY crontab /etc/crontabs/root

RUN pip install -r requeriments.txt

CMD ["supercronic", "crontab"]
#CMD ["crond", "-f", "-d", "8"]