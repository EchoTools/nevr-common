# Makefile for generating protocol buffer files in multiple languages
PROTO_DIRS := api rtapi gameapi
PROTO_FILES := $(foreach dir,$(PROTO_DIRS),$(wildcard $(dir)/*.proto))
PY_FILES := $(PROTO_FILES:.proto=_pb2.py)
GRPC_PY_FILES := $(PROTO_FILES:.proto=_pb2_grpc.py)
GO_FILES := $(PROTO_FILES:.proto=.pb.go)
CSHARP_FILES := $(PROTO_FILES:.proto=.cs)

# Generate -I flags for each proto dir
PROTO_INCLUDES := $(foreach dir,$(PROTO_DIRS),-I./$(dir))

all: $(PY_FILES) $(GRPC_PY_FILES) $(GO_FILES) $(CSHARP_FILES)

%_pb2.py: %.proto
	protoc $(PROTO_INCLUDES) --python_out=$(dir $<) $<

%_pb2_grpc.py: %.proto
	python -m grpc_tools.protoc $(PROTO_INCLUDES) --python_out=$(dir $<) --grpc_python_out=$(dir $<) $<

%.pb.go: %.proto
	go generate ./$(dir $<)

%.cs: %.proto
	protoc $(PROTO_INCLUDES) --csharp_out=$(dir $<) $<

clean:
	rm -f $(PY_FILES) $(GRPC_PY_FILES) $(GO_FILES) $(CSHARP_FILES)

.PHONY: all clean
