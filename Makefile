gracc-collector: *.go
	go build -ldflags "-X main.build_date=`date -u +%Y%m%d.%H%M%S` -X main.build_ref=RELEASE" -o gracc-collector

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


clean:
	rm -f gracc-collector
