FROM centos:7.8.2003 

RUN yum update -y && \
    yum install -y wget hostnamectl && \
    wget $(curl https://api.github.com/repos/mikefarah/yq/releases/tags/3.4.1 | egrep yq_linux_amd64 | cut -d \" -f4) || [ "$?" == 4 ] && \
    chmod +x yq_linux_amd64 && \
    mv yq_linux_amd64 /usr/local/bin/yq

ENV container docker

VOLUME [ "/sys/fs/cgroup" ]

CMD ["/usr/sbin/init"]

COPY ./artifacts/nodeagent /

     



