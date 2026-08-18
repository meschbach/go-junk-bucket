BENCHSTAT := $(shell which benchstat 2>/dev/null || echo $(HOME)/go/bin/benchstat)

pump: bin
	go build -o bin/pump ./cli/pump/main.go

bin:
	mkdir -p bin

clean:
	rm -fR bin

test:
	go test -count 1 --timeout 10s ./pkg/... ./sub/...

lint:
	golangci-lint run ./pkg/... ./sub/...

bench:
	@mkdir -p bench-results
	go test -bench Benchmark -benchmem -count 6 -timeout 120s ./pkg/emitter/ | tee bench-results/emitter.txt
	$(BENCHSTAT) bench-results/emitter.txt

bench-baseline: bench
	cp bench-results/emitter.txt bench-results/emitter-baseline.txt

bench-compare:
	@test -f bench-results/emitter-baseline.txt || { echo "No baseline found. Run 'make bench-baseline' first."; exit 1; }
	@mkdir -p bench-results
	go test -bench Benchmark -benchmem -count 6 -timeout 120s ./pkg/emitter/ | tee bench-results/emitter.txt
	$(BENCHSTAT) bench-results/emitter-baseline.txt bench-results/emitter.txt
