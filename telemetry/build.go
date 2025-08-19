package telemetry

//go:generate protoc -I. -I.. --go_out=. --go_opt=paths=source_relative telemetry.proto
