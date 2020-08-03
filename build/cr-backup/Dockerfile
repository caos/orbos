FROM cockroachdb/cockroach:v20.1.3

RUN apt-get -y update && apt-get -y install curl wget

RUN wget https://storage.googleapis.com/oauth2l/latest/linux_amd64.tgz
RUN tar zxvf linux_amd64.tgz
RUN mv linux_amd64/oauth2l /usr/local/bin/oauth2l && rm -rf linux_amd64

COPY build/scripts/backup-cockroach.sh /scripts/backup.sh
RUN chmod +x /scripts/backup.sh

COPY build/scripts/restore-cockroach.sh /scripts/restore.sh
RUN chmod +x /scripts/restore.sh

COPY build/scripts/clean-cockroach.sh /scripts/clean.sh
RUN chmod +x /scripts/clean.sh

ENTRYPOINT [ "/cockroach" ]