FROM opensciencegrid/osg-wn:3.3-el7

LABEL name="OSG GRACC Collector"
LABEL build-date=20170516-1513

# install gracc from RPM
RUN yum -y --enablerepo=osg-development install gracc-collector && \
    yum clean all

ENV GRACC_ADDRESS 0.0.0.0
ENV GRACC_PORT 8080

EXPOSE 8080
CMD ["/usr/bin/gracc-collector"]
