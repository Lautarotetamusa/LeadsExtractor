FROM python:3.10.13-alpine3.18
ENV PYTHONUNBUFFERED=1

RUN apk update
RUN set -x \
    && apk add --no-cache supercronic shadow
RUN apk add chromium chromium-chromedriver xvfb

#Poner el timezone correcto
ENV TZ="America/Mexico_City"
RUN date
RUN cp /usr/share/zoneinfo/$TZ etc/localtime

WORKDIR app

COPY requirements.txt src/
RUN pip install -r src/requirements.txt

COPY . . 

EXPOSE 80

COPY entrypoint.sh .
RUN chmod +x /app/entrypoint.sh
CMD ["/app/entrypoint.sh"]
