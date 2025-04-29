FROM ubuntu:20.04

RUN dpkg --add-architecture i386 && \
    apt-get update && \
    apt-get install -y \
    libstdc++6:i386 \
    ca-certificates

RUN useradd -ms /bin/bash hltv

WORKDIR /home/hltv
COPY hltv .
COPY filesystem_stdio.so .
COPY proxy.so .
COPY libsteam_api.so /usr/lib
COPY core.so .
COPY steamclient.so .
COPY libsteam.so .steam/sdk32/

RUN mkdir -p /home/hltv/cstrike

RUN chmod +x ./hltv && \
    chown -R hltv:hltv /home/hltv

USER hltv

VOLUME ["/home/hltv/cstrike"]

ENV LD_LIBRARY_PATH=./

ENTRYPOINT ["./hltv"]
CMD ["+connect", "127.0.0.1:27015", "-port", "1337", "+record", "demoname", "+exec", "hltv.cfg"]
