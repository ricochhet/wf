CUSTOM=-X 'main.buildDate=$(shell date)' -X 'main.gitHash=$(shell git rev-parse --short HEAD)' -X 'main.buildOn=$(shell go version)'
LDFLAGS=$(CUSTOM) -w -s -extldflags=-static

GO_BUILD=go build -trimpath -ldflags "$(LDFLAGS)"
BUILD_OUTPUT=build
ASSET_PATH=assets

# pkg
PKG_PATH=./pkg

# cmd
DBMOD_PATH=./cmd/dbmod
DBMOD_ASSETS_PATH=./cmd/dbmod/assets
GPM_PATH=./cmd/gpm

# linters
CHECKNEWLINES_PATH=./linters/checknewlines
CHECKRECV_PATH=./linters/checkrecv

.PHONY: all
all: dbmod-linux dbmod-linux-arm64 dbmod-darwin dbmod-darwin-arm64 dbmod-windows

.PHONY: fmt
fmt:
	gofumpt -l -w -extra .

.PHONY: tidy
tidy:
#	go get -u ./...
	@echo "[pkg] tidy"
	cd $(PKG_PATH) && go mod tidy
	@echo "[cmd] tidy"
	cd $(DBMOD_PATH) && go mod tidy
	cd $(GPM_PATH) && go mod tidy
# linters
	@echo "[linters] tidy"
	cd $(CHECKNEWLINES_PATH) && go mod tidy
	cd $(CHECKRECV_PATH) && go mod tidy
# don't include submodules

.PHONY: update
update:
	@echo "[pkg] tidy"
	cd $(PKG_PATH) && go get -u ./...
	@echo "[cmd] tidy"
	cd $(DBMOD_PATH) && go get -u ./...
	cd $(GPM_PATH) && go get -u ./...
# linters
	@echo "[linters] tidy"
	cd $(CHECKNEWLINES_PATH) && go get -u ./...
	cd $(CHECKRECV_PATH) && go get -u ./...
# don't include submodules

.PHONY: lint
lint: fmt
# golangci-lint cache clean
	@echo "[pkg] golangci-lint"
	cd $(PKG_PATH) && golangci-lint run ./... --fix
	@echo "[cmd] golangci-lint"
	cd $(DBMOD_PATH) && golangci-lint run --fix
	cd $(GPM_PATH) && golangci-lint run --fix
# linters
	@echo "[linters] golangci-lint"
	cd $(CHECKNEWLINES_PATH) && golangci-lint run --fix
	cd $(CHECKRECV_PATH) && golangci-lint run --fix
# don't include submodules / thirdparty tools

.PHONY: test
test:
	go test ./...

.PHONY: deadcode
deadcode:
	deadcode ./...

.PHONY: syso
syso:
	windres $(DBMOD_PATH)/app.rc -O coff -o $(DBMOD_PATH)/app.syso
	windres $(GPM_PATH)/app.rc -O coff -o $(GPM_PATH)/app.syso

.PHONY: download-rcedit
download-rcedit:
	@version=$(version); \
	[ -z "$$version" ] && version="v2.0.0"; \
	url="https://github.com/electron/rcedit/releases/download/$$version/rcedit-x64.exe"; \
	echo "Downloading $$url..."; \
	curl -L -o rcedit-x64.exe $$url

.PHONY: png-to-ico
png-to-ico:
	magick $(ASSET_PATH)/win-icon.png -background none -define icon:auto-resize=256,128,64,48,32,16 $(ASSET_PATH)/win-icon.ico

.PHONY: dbmod-linux
dbmod-linux: fmt
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GO_BUILD) -o $(BUILD_OUTPUT)/dbmod-linux $(DBMOD_PATH)

#### dbmod

.PHONY: dbmod-linux-arm64
dbmod-linux-arm64: fmt
	CGO_ENABLED=1 GOOS=linux GOARCH=arm64 $(GO_BUILD) -o $(BUILD_OUTPUT)/dbmod-linux-arm64 $(DBMOD_PATH)

.PHONY: dbmod-darwin
dbmod-darwin: fmt
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 $(GO_BUILD) -o $(BUILD_OUTPUT)/dbmod-darwin $(DBMOD_PATH)

.PHONY: dbmod-darwin-arm64
dbmod-darwin-arm64: fmt
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 $(GO_BUILD) -o $(BUILD_OUTPUT)/dbmod-darwin-arm64 $(DBMOD_PATH)

.PHONY: dbmod-windows-build
dbmod-windows-build: fmt
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 $(GO_BUILD) -o $(BUILD_OUTPUT)/dbmod.exe $(DBMOD_PATH)

.PHONY: dbmod-windows-postbuild
dbmod-windows-postbuild: fmt
	cp -r $(DBMOD_ASSETS_PATH) $(BUILD_OUTPUT)/
# icons
	rcedit-x64 $(BUILD_OUTPUT)/dbmod.exe --set-icon $(ASSET_PATH)/win-icon.ico

.PHONY: dbmod-windows
dbmod-windows: dbmod-windows-build dbmod-windows-postbuild
	cp build/dbmod.exe dbmod.exe

#### gpm

.PHONY: gpm-linux
gpm-linux: fmt
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GO_BUILD) -o $(BUILD_OUTPUT)/gpm-linux $(GPM_PATH)

.PHONY: gpm-linux-arm64
gpm-linux-arm64: fmt
	CGO_ENABLED=1 GOOS=linux GOARCH=arm64 $(GO_BUILD) -o $(BUILD_OUTPUT)/gpm-linux-arm64 $(GPM_PATH)

.PHONY: gpm-darwin
gpm-darwin: fmt
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 $(GO_BUILD) -o $(BUILD_OUTPUT)/gpm-darwin $(GPM_PATH)

.PHONY: gpm-darwin-arm64
gpm-darwin-arm64: fmt
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 $(GO_BUILD) -o $(BUILD_OUTPUT)/gpm-darwin-arm64 $(GPM_PATH)

.PHONY: gpm-windows-build
gpm-windows-build: fmt
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 $(GO_BUILD) -o $(BUILD_OUTPUT)/gpm.exe $(GPM_PATH)

.PHONY: gpm-windows-postbuild
gpm-windows-postbuild: fmt
# icons
	rcedit-x64 $(BUILD_OUTPUT)/gpm.exe --set-icon $(ASSET_PATH)/win-icon.ico

.PHONY: gpm-windows
gpm-windows: gpm-windows-build gpm-windows-postbuild
	cp build/gpm.exe gpm.exe

#### linters

.PHONY: checknewlines-windows
checknewlines-windows: fmt
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 $(GO_BUILD) -o $(BUILD_OUTPUT)/checknewlines.exe $(CHECKNEWLINES_PATH)

.PHONY: checkrecv-windows
checkrecv-windows: fmt
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 $(GO_BUILD) -o $(BUILD_OUTPUT)/checkrecv.exe $(CHECKRECV_PATH)

.PHONY: linters-windows
linters-windows: checknewlines-windows checkrecv-windows

.PHONY: clean-dbmod
clean-dbmod:
	rm -f dbmod-linux dbmod-linux-arm64 dbmod-darwin dbmod-darwin-arm64 dbmod.exe

.PHONY: clean-gpm
clean-gpm:
	rm -f gpm-linux gpm-linux-arm64 gpm-darwin gpm-darwin-arm64 gpm.exe