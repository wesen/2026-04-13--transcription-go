.PHONY: build run test clean

INPUT ?= 
OUTPUT ?= ./out

build:
	go build -o transcribe ./cmd/transcribe

run: build
	./transcribe --input $(INPUT) --output-dir $(OUTPUT) --format srt,db

test:
	go test ./... -count=1

clean:
	rm -f transcribe
	rm -rf out/
