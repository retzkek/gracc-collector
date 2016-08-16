FROM local/golang:centos7

# Build & Test
RUN mkdir -p /gopath/src/github.com/opensciencegrid/gracc-collector
ADD . /gopath/src/github.com/opensciencegrid/gracc-collector

WORKDIR /gopath/src/github.com/opensciencegrid/gracc-collector
CMD make test && make rpm
