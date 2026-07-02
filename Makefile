.PHONY: build run tidy clean

NIX_DEPS = go gcc pkg-config xorg.libX11 xorg.libXxf86vm xorg.libXcursor \
	xorg.libXrandr xorg.libXinerama xorg.libXi libglvnd mesa

build:
	nix-shell -p $(NIX_DEPS) --run "go build -o bin/trackapp ."

run: build
	./bin/trackapp

tidy:
	nix-shell -p go --run "go mod tidy"

clean:
	rm -rf bin/trackapp