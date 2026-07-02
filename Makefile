.PHONY: build run tidy clean test darwin-arm64 darwin-amd64 windows

NIX_DEPS = go gcc pkg-config xorg.libX11 xorg.libXxf86vm xorg.libXcursor \
	xorg.libXrandr xorg.libXinerama xorg.libXi libglvnd mesa

build:
	nix-shell -p $(NIX_DEPS) --run "go build -o bin/trackapp ."

run: build
	./bin/trackapp

test:
	nix-shell -p go --run "go test ./..."

tidy:
	nix-shell -p go --run "go mod tidy"

clean:
	rm -rf bin/trackapp bin/trackapp-darwin-* bin/trackapp.exe

# macOS binaries must be built on macOS (Fyne/CGO needs the Xcode toolchain).
darwin-arm64:
	go build -o bin/trackapp-darwin-arm64 .

darwin-amd64:
	GOARCH=amd64 go build -o bin/trackapp-darwin-amd64 .

windows:
	nix-shell -p go --run "GOOS=windows GOARCH=amd64 go build -o bin/trackapp.exe ."