FROM metabrainz/docker-anon-ftp

RUN apt-get update \
    && apt-get install -y --no-install-recommends p7zip-full p7zip-rar wget \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

ENV ANON_ROOT=/centos \
    PASV_MAX_PORT=65500 \
    PASV_MIN_PORT=65500 \
    MAX_CLIENTS=1

WORKDIR $ANON_ROOT
RUN wget -q http://mirror.init7.net/centos/7.7.1908/isos/x86_64/CentOS-7-x86_64-Minimal-1908.iso
RUN 7z x CentOS-7-x86_64-Minimal-1908.iso && rm CentOS-7-x86_64-Minimal-1908.iso
RUN ls -la images/pxeboot
RUN pwd