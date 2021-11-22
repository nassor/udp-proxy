SHELL=/bin/bash -o pipefail

.PHONY: build
build:
	go build -o bin/proxy ./cmd/udp-proxy

.PHONY: run
run: build 
	./bin/proxy

.PHONY: load
load:
	go build -o bin/load-generator ./cmd/load
	./bin/load-generator
	
.PHONY: pprof-memory
pprof-memory:
	go tool pprof -png http://localhost:3000/debug/pprof/allocs > ./test/pprof/allocs.png
	go tool pprof -png http://localhost:3000/debug/pprof/heap > ./test/pprof/heap.png

.PHONY: pprof-cpu
pprof-cpu:
	go tool pprof -png -seconds=15 http://localhost:3000/debug/pprof/profile > ./test/pprof/cpu.png

.PHONY: pprof
pprof: pprof-memory pprof-cpu


.PHONY: trace
trace: benchmark
	go tool trace ./test/trace.out

.PHONY: gcvis
# it needs to have gcvis installed
# https://github.com/davecheney/gcvis
gcvis: build
	gcvis ./bin/proxy