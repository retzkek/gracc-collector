gratia2-collector: *.go
	go build -ldflags "-X main.buildDate=`date -u +%Y%m%d.%H%M%S` -X main.commit=`git rev-parse --verify HEAD --short`"

test:
	go test -v

clean:
	rm -f gratia2-collector
