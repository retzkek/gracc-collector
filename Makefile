.PHONY: help scratch run test rpm docker with-docker docker-setup docker-build docker-clean clean

gracc-collector: *.go
	go build -ldflags "-X main.build_date=`date -u +%Y%m%d.%H%M%S` -X main.build_ref=RELEASE" -o gracc-collector

help:
	 @echo  'Build Targets'
	 @echo  '  scratch         - build scratch executable with race checking'
	 @echo  '  run             - build temporary exectutable and run it'
	 @echo  '  test            - build and run tests via "go test"'
	 @echo  '  rpm             - build RPM from HEAD'
	 @echo  '  docker          - build minimal deployable docker image'
	 @echo  '  with-docker     - build & test in docker image.'
	 @echo  '                    if successful, saves exectuable and RPM in targets/'
	 @echo  ''
	 @echo  'Clean Targets'
	 @echo  '  docker-clean    - cleans up docker images and networks from with-docker'
	 @echo  '  clean           - cleans up executables'
	 @echo  ''

scratch: *.go
	go build -ldflags "-X main.build_date=`date -u +%Y%m%d.%H%M%S` -X main.build_ref=`git rev-parse --verify HEAD --short`" -race -o gracc-collector.scratch

run:
	go build -ldflags "-X main.build_date=`date -u +%Y%m%d.%H%M%S`" -o gracc.run -race && ./gracc.run; rm -f gracc.run

test:
	# test gracc library
	cd gracc; go test
	# test main
	go test

rpm:
	git archive --prefix gracc-collector/ --output $(HOME)/rpmbuild/SOURCES/gracc-collector.tar.gz HEAD
	rpmbuild -ba gracc-collector.spec

docker:
	CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags "-X main.build_date=`date -u +%Y%m%d.%H%M%S` -X main.build_ref=RELEASE -w" -o gracc-collector
	docker build -f Dockerfile.deploy -t opensciencegrid/gracc-collector .

with-docker: | docker-setup docker-build docker-rpmtest docker-clean

docker-setup:
	docker network create gracc
	docker run -d --network gracc --name rabbit rabbitmq

docker-build:
	docker build -t opensciencegrid/gracc-collector-test .
	docker run -it --network gracc --name gracc -e GRACC_AMQP_HOST=rabbit opensciencegrid/gracc-collector-test
	docker cp gracc:/root/rpmbuild/RPMS/ ./target/
	docker cp gracc:/root/rpmbuild/BUILD/gracc-collector/gracc-collector ./target/

docker-rpmtest:
	docker build -f Dockerfile.rpmtest -t opensciencegrid/gracc-rpmtest .
	docker run --privileged -d --network gracc --name gracc-rpm -v /sys/fs/cgroup:/sys/fs/cgroup:ro opensciencegrid/gracc-rpmtest
	-docker exec gracc-rpm /usr/bin/systemctl status gracc-collector
	-docker exec gracc-rpm docker exec gracc-rpm cat /var/log/gracc/gracc-collector.log
	docker stop gracc-rpm

docker-clean:
	-docker rm -f rabbit gracc gracc-rpm
	-docker network rm gracc

clean:
	rm -f gracc-collector gracc-collector.scratch gracc.run
