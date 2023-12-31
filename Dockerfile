FROM python:3.10.13-alpine3.18
ENV PYTHONUNBUFFERED=1

RUN apk update
RUN set -x \
    && apk add --no-cache supercronic shadow
RUN apk add chromium chromium-chromedriver xvfb

WORKDIR app

COPY requirements.txt src/
RUN pip install -r src/requirements.txt
COPY src/ src/
COPY server/ .
COPY messages/ messages/

COPY credentials.json .

COPY crontab.sh .
COPY src/start_xvfb.sh .
COPY mapping.json .
COPY token.json .
COPY scraper_mapping.json .
COPY main.py .
COPY auth.py .
COPY gmail_first_run.py .
COPY sended.json .
COPY last_sended.txt .

EXPOSE 80

COPY entrypoint.sh .
RUN chmod +x /app/entrypoint.sh
CMD ["/app/entrypoint.sh"]
