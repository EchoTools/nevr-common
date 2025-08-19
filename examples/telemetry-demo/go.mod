module telemetry-demo

go 1.24.3

replace github.com/echotools/nevr-common/v3 => ../..

require (
	github.com/echotools/nevr-common/v3 v3.0.0-00010101000000-000000000000
	google.golang.org/protobuf v1.36.7
)

require (
	github.com/klauspost/compress v1.18.0 // indirect
	nhooyr.io/websocket v1.8.17 // indirect
)
