gracc: *.go
	go build -ldflags "-X main.build_date=`date -u +%Y%m%d.%H%M%S` -X main.build_ref=`git rev-parse --verify HEAD --short`" -o gracc

scratch: *.go
	go build -ldflags "-X main.build_date=`date -u +%Y%m%d.%H%M%S`" -race -o gracc.scratch

run:
	go build -ldflags "-X main.build_date=`date -u +%Y%m%d.%H%M%S`" -o gracc.run -race && ./gracc.run; rm -f gracc.run

test:
	rm -rf /tmp/gracc.test
	go test -v -race

servertest:
	# send ping
	curl -i http://localhost:8080/gratia-servlets/rmi\?command\=update\&arg1\=xxx\&from\=localhost\&bundlesize\=1
	@echo
	# send test bundle
	curl -i http://localhost:8080/gratia-servlets/rmi\?command\=update\&from\=localhost\&bundlesize\=10 --data-urlencode arg1@test.bundle
	@echo
	# send bad requests
	curl -i http://localhost:8080/gratia-servlets/rmi
	@echo
	curl -i http://localhost:8080/gratia-servlets/rmi\?command\=noop
	@echo
	curl -i http://localhost:8080/gratia-servlets/rmi\?command\=update
	@echo

clean:
	rm -f gracc
