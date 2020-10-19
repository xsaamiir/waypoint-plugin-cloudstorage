PLUGIN_NAME=gcs

all: protos build

protos:
	@echo ""
	@echo "Building Protos"

	protoc -I . --go_opt=paths=source_relative --go_out=. ./registry/output.proto

.PHONY: build
build:
	@echo ""
	@echo "Compiling Plugin"

	go build -o ./bin/waypoint-plugin-${PLUGIN_NAME} ./main.go

.PHONY: install
install: build
	@echo ""
	@echo "Installing Plugin"

	cp ./bin/waypoint-plugin-${PLUGIN_NAME} ${HOME}/.config/waypoint/plugins/
