createmigration:
		migrate create -ext sql -dir sql/migrations -seq init

migrate:
		migrate -path=sql/migrations -database "mysql://root:root@tcp(127.0.0.1:3306)/chat_app" -verbose up

migratedown:
		migrate -path=sql/migrations -database "mysql://root:root@tcp(127.0.0.1:3306)/chat_app" -verbose drop

grpc:
		protoc --experimental_allow_proto3_optional --go_out=. --go-grpc_out=. proto/chat.proto

.PHONY: createmigration migrate migratedown grpc
