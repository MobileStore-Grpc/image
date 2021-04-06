import-proto:
	mkdir -p proto/import && \
	go mod download && \
	go list -f "proto/import/{{ .Path }}/proto" -m all \
	| grep proto/import/github.com/MobileStore-Grpc/product/proto | xargs -L1 dirname | sort | uniq | xargs mkdir -p && \
	go list -f "{{ .Dir }}/proto proto/import/{{ .Path }}/proto" -m all \
  	| grep proto/import/github.com/MobileStore-Grpc/product/proto | xargs -L1 -- ln -s

delete-proto-import: 
	find  proto/import -type l -delete && \
	find proto/import -type d -empty -delete

gen:
	protoc -I=proto/ -I=proto/import/github.com/MobileStore-Grpc/product/proto/ \
	--go_out=. --go_opt=module=github.com/MobileStore-Grpc/image \
	--go-grpc_out=. --go-grpc_opt=module=github.com/MobileStore-Grpc/image \
	--grpc-gateway_out=. --grpc-gateway_opt=module=github.com/MobileStore-Grpc/image \
	--openapiv2_out=swagger \
	proto/*.proto

clean:
	rm -r pb/*.go swagger/*

server:
	go run cmd/server/main.go --port 8080

rest:
	go run cmd/server/main.go --port 8082 --type rest --endpoint 0.0.0.0:8080

client:
	go run cmd/client/main.go --address localhost:9020


build-image:
	docker build -t mobilestore-image:v1.0.0 .

run:
	docker run -d --name image -p 9020:8020 mobilestore-image:v1.0.0