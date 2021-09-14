pump: bin
	go build -o bin/pump ./cli/pump/main.go

bin:
	mkdir -p bin

clean:
	rm -fR bin