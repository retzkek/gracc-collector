gracc: *.go
	go build -ldflags "-X main.buildDate=`date -u +%Y%m%d.%H%M%S` -X main.commit=`git rev-parse --verify HEAD --short`"

run:
	go build -ldflags "-X main.buildDate=`date -u +%Y%m%d.%H%M%S`" -o gracc.run && ./gracc.run; rm -f gracc.run

test:
	go test -v

clean:
	rm -f gracc
