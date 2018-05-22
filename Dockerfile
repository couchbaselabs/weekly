FROM alpine:3.5

MAINTAINER Pavel Paulau <pavel@couchbase.com>

EXPOSE 9009

ENV CB_HOST ""
ENV CB_PASS ""

COPY app app
COPY weekly /usr/local/bin/weekly

CMD ["weekly"]
