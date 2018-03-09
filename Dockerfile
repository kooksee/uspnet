FROM ubuntu:16.04

RUN rm -rf /app && mkdir /app && mkdir /kdata
COPY main /app/server
WORKDIR /app

EXPOSE 46380
EXPOSE 46381
EXPOSE 46382
EXPOSE 46383

VOLUME /kdata

CMD ["s"]
ENTRYPOINT ["/app/server","--home","/kdata"]