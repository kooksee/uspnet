FROM ubuntu:16.04

RUN rm -rf /app && mkdir /app && mkdir /kdata
COPY main /app/kchain
WORKDIR /app

EXPOSE 8080

VOLUME /kdata

CMD ["node"]
ENTRYPOINT ["/app/kchain","--home","/kdata"]