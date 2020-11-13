FROM cockroachdb/cockroach:v20.1.8

RUN apt-get -y update && apt-get -y install curl wget

RUN wget https://storage.googleapis.com/oauth2l/latest/linux_amd64.tgz
RUN tar zxvf linux_amd64.tgz
RUN mv linux_amd64/oauth2l /usr/local/bin/oauth2l && rm -rf linux_amd64

COPY build/scripts/backup-cockroach.sh /scripts/backup.sh
RUN chmod +x /scripts/backup.sh

COPY build/scripts/restore-cockroach.sh /scripts/restore.sh
RUN chmod +x /scripts/restore.sh

COPY build/scripts/clean-db-cockroach.sh /scripts/clean-db.sh
RUN chmod +x /scripts/clean-db.sh
COPY build/scripts/clean-migration-cockroach.sh /scripts/clean-migration.sh
RUN chmod +x /scripts/clean-migration.sh
COPY build/scripts/clean-user-cockroach.sh /scripts/clean-user.sh
RUN chmod +x /scripts/clean-user.sh

ENTRYPOINT [ "/cockroach" ]
