FROM ubuntu:16.04

RUN rm -rf /app && mkdir /app && mkdir /kdata
COPY main /app/server
WORKDIR /app

EXPOSE 8080

VOLUME /kdata

CMD ["s"]
ENTRYPOINT ["/app/server","--home","/kdata"]