module github.com/opensoba/opensoba/internal/tools/protoc-gen-sabo-rate-limit

go 1.19

replace github.com/opensoba/opensoba/proto-public => ../../../proto-public

require (
	github.com/opensoba/opensoba/proto-public v0.0.0-00010101000000-000000000000
	google.golang.org/protobuf v1.30.0
)

require github.com/google/go-cmp v0.5.9 // indirect
