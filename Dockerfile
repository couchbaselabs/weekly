FROM alpine:3.7

MAINTAINER Pavel Paulau <pavel@couchbase.com>

EXPOSE 9009

ENV CB_HOST ""
ENV CB_PASS ""
ENV CB_USER ""

COPY app app
COPY weekly /usr/local/bin/weekly

CMD ["weekly"]
