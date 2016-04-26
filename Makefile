_build/shorturl:
	rm -rf _build
	GOOS=linux GOARCH=amd64 go build -o _build/shorturl cmd/shorturl-server/main.go
	git archive HEAD content|tar x -C _build
	tar zcf shorturl.tar.gz -C _build .

.PHONY: clean
clean:
	rm -rf _build

run:
	go run cmd/shorturl-server/main.go content

deploy:
	rsync -av _build/ lakka:apps/shorturl-go/
	ssh lakka supervisorctl restart shorturl
	curl https://yx.fi/1

all: _build/shorturl
