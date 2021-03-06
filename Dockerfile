FROM alpine:latest
EXPOSE 3000
RUN apk --no-cache add curl openssh-client ca-certificates tzdata

# Set local time (for cronjob sense)
RUN cp /usr/share/zoneinfo/Europe/Berlin /etc/localtime && \
echo "Europe/Berlin" > /etc/timezone

ADD telegram /app/

WORKDIR /app

ENTRYPOINT ["/app/telegram"]