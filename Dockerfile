FROM python:alpine
ENV PYTHONUNBUFFERED=1

RUN apk update
RUN set -x \
    && apk add --no-cache supercronic shadow
RUN apk add chromium chromium-chromedriver xvfb

WORKDIR app

COPY src/ src/
COPY requeriments.txt src/

COPY crontab.sh .
COPY credentials.json .
COPY token.json .
COPY src/start_xvfb.sh .
COPY mapping.json .
COPY main.py .

RUN pip install -r src/requeriments.txt

CMD ["supercronic", "crontab.sh"]