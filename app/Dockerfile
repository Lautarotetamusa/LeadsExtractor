FROM python:3.10.13-alpine3.18
ENV PYTHONUNBUFFERED=1

RUN apk update
RUN set -x \
    && apk add --no-cache supercronic shadow
RUN apk add chromium chromium-chromedriver xvfb

WORKDIR app

COPY ./app/requirements.txt src/
RUN pip install -r src/requirements.txt

COPY ./app/ . 
COPY ../messages ../messages

EXPOSE 80

COPY ./app/entrypoint.sh .
RUN chmod +x /app/entrypoint.sh
CMD ["/app/entrypoint.sh"]
