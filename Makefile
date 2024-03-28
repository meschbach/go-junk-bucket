pump: bin
	go build -o bin/pump ./cli/pump/main.go

bin:
	mkdir -p bin

clean:
	rm -fR bin

test:
	go test -count 1 --timeout 10s ./pkg/... ./sub/...
