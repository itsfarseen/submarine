run:
	go run app/main.go

codegen:
	go run cmd/codegen/main.go

scan-spec-versions:
	go run cmd/scan-spec-versions/main.go

scan-spec-types:
	go run cmd/scan-spec-types/main.go
