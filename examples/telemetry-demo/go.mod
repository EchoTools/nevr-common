module rtapi-demo

go 1.24.3

replace github.com/echotools/nevr-common => ../..

require (
	github.com/echotools/nevr-common v0.0.0-00010101000000-000000000000
	google.golang.org/protobuf v1.36.9
)

require github.com/klauspost/compress v1.18.0 // indirect
