run:
	go run main.go
migrate_up:
	migrate -path internal/db/migrations -database postgres://sayyidmuhammad:root@localhost:5432/postgres?sslmode=disable -verbose up

migrate_down:
	migrate -path internal/db/migrations -database postgres://sayyidmuhammad:root@localhost:5432/postgres?sslmode=disable -verbose down

migrate_force:
	migrate -path internal/db/migrations -database postgres://sayyidmuhammad:root@localhost:5432/postgres?sslmode=disable -verbose force 1

migrate_file:
	migrate create -ext sql -dir internal/db/migrations -seq create_table

swag-gen:
	export PATH="$PATH:$(go env GOPATH)/bin"
	swag init -g pkg/api/v1/router.go -o ./docs