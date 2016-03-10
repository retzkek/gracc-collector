gracc: *.go
	go build -ldflags "-X main.build_date=`date -u +%Y%m%d.%H%M%S` -X main.build_ref=`git rev-parse --verify HEAD --short`"

run:
	go build -ldflags "-X main.build_date=`date -u +%Y%m%d.%H%M%S`" -o gracc.run && ./gracc.run; rm -f gracc.run

test:
	rm -rf /tmp/gracc.test
	go test -v

servertest:
	# send ping
	curl http://localhost:8080/gratia-servlets/rmi\?command\=update\&arg1\=xxx\&from\=localhsot\&bundlesize\=1
	# send test bundle
	curl http://localhost:8080/gratia-servlets/rmi\?command\=update\&from\=localhost\&bundlesize\=10 --data-urlencode arg1@test.bundle

clean:
	rm -f gracc
