gen:
	protoc --proto_path=proto proto/*.proto --go_out=plugins=grpc:pb --grpc-gateway_out=:pb --swagger_out=:swagger

clean:
	rm pb/*.go

server:
	go run cmd/server/main.go -port 8080

client:
	go run cmd/client/main.go -address 0.0.0.0:8080

rest:
	# go run cmd/server/main.go -port 8081 -type rest
	# improve
	go run cmd/server/main.go -port 8081 -type rest -endpoint  0.0.0.0:8080

test:
	go test -cover -race ./...

cert:
	cd cert; sh gen.sh; cd ..

# Nginx Load Balance Test Start
server1:
	go run cmd/server/main.go -port 9001

server2:
	go run cmd/server/main.go -port 9002

server1-tls:
	go run cmd/server/main.go -port 9001 -tls

server2-tsl:
	go run cmd/server/main.go -port 9002 -tls

client-tls:
	go run cmd/client/main.go -address 0.0.0.0:8080 -tls
# Nginx Load Balance Test End

.PHONY: gen clean server client test cert
