.PHONY: migrate-up migrate-down sqlc server proto server tss init proto

include .env

PATH_CURRENT := $(shell pwd)
PATH_BUILT := $(PATH_CURRENT)/build/server
GIT_COMMIT_LOG := $(shell git log --oneline -1 HEAD)

init: 
	git submodule init
	git submodule update

migrate-up:
	migrate -path db/migration -database "${DB_URL}" -verbose up

migrate-down:
	migrate -path db/migration -database "${DB_URL}" -verbose down

sqlc:
	sqlc generate

server:
	go run main.go
tss:
	go build -o cmd/tss main.go

node-1-test:
	go run main.go --config config/id-10002-input.yaml  --keystore ./node.test/keystore/2

proto:
	rm -f pb/*.go
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
	--go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	--descriptor_set_out descriptor.pb \
	proto/*.proto
