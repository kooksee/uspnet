FROM ubuntu:16.04

RUN rm -rf /app && mkdir /app && mkdir /kdata
COPY main /app/server
WORKDIR /app

EXPOSE 46383

VOLUME /kdata

CMD ["s"]
ENTRYPOINT ["/app/server","--home","/kdata"]