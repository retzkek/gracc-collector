FROM centos:7

# install gracc from RPM
ADD ./target/RPMS/x86_64/ /rpms
RUN yum -y install $(ls -r /rpms/gracc-collector-*.el7.centos.x86_64.rpm | head -n 1)

ENV GRACC_ADDRESS 0.0.0.0
ENV GRACC_PORT 8080

EXPOSE 8080
CMD ["/usr/bin/gracc-collector"]
