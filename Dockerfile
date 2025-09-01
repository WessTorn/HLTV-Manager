FROM ubuntu:20.04

RUN dpkg --add-architecture i386 && \
    apt-get update && \
    apt-get install -y \
    libstdc++6:i386 \
    ca-certificates \
    tzdata && \
    rm -rf /var/lib/apt/lists/*

ENV TZ=Europe/Moscow

RUN ln -sf /usr/share/zoneinfo/$TZ /etc/localtime && \
    echo $TZ > /etc/timezone && \
    dpkg-reconfigure -f noninteractive tzdata

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
RUN mkdir -p /home/hltv/tfc
RUN mkdir -p /home/hltv/dod
RUN mkdir -p /home/hltv/dmc
RUN mkdir -p /home/hltv/gearbox
RUN mkdir -p /home/hltv/ricochet
RUN mkdir -p /home/hltv/valve
RUN mkdir -p /home/hltv/czero
RUN mkdir -p /home/hltv/czeror
RUN mkdir -p /home/hltv/bshift
RUN mkdir -p /home/hltv/cstrike_beta


RUN chmod +x ./hltv && \
    chown -R hltv:hltv /home/hltv

USER hltv

VOLUME ["/home/hltv/cstrike"]
VOLUME ["/home/hltv/tfc"]
VOLUME ["/home/hltv/dod"]
VOLUME ["/home/hltv/dmc"]
VOLUME ["/home/hltv/gearbox"]
VOLUME ["/home/hltv/ricochet"]
VOLUME ["/home/hltv/valve"]
VOLUME ["/home/hltv/czero"]
VOLUME ["/home/hltv/czeror"]
VOLUME ["/home/hltv/bshift"]
VOLUME ["/home/hltv/cstrike_beta"]

ENV LD_LIBRARY_PATH=./

ENTRYPOINT ["./hltv"]
CMD ["+connect", "127.0.0.1:27015", "-port", "1337", "+record", "demoname", "+exec", "hltv.cfg"]
