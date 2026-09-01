APP := Highway
APP_ID := com.herrmann.highway
DIST := dist
ICON_SVG := assets/highway.svg
ICON_PNG := assets/highway.png
FYNE := go run fyne.io/tools/cmd/fyne@latest

.PHONY: macos macos-service linux icon clean test

macos: icon
	mkdir -p $(DIST)
	rm -rf $(DIST)/$(APP).app $(APP).app
	$(FYNE) package -os darwin -name $(APP) -appID $(APP_ID) -icon $(ICON_PNG)
	mv $(APP).app $(DIST)/$(APP).app

macos-service:
	zsh packaging/macos/install-service.sh

linux: icon
	@if [ "$$(go env GOOS)" != "linux" ] || [ "$$(go env GOARCH)" != "amd64" ]; then \
		echo "Execute 'make linux' em uma maquina Linux x86_64."; exit 1; \
	fi
	mkdir -p $(DIST)
	go build -o $(DIST)/highway .
	cp packaging/highway.desktop $(DIST)/highway.desktop
	cp $(ICON_PNG) $(DIST)/highway.png
	tar -czf $(DIST)/highway-linux-amd64.tar.gz -C $(DIST) highway highway.desktop highway.png

icon:
	@if command -v sips >/dev/null 2>&1; then \
		sips -s format png $(ICON_SVG) --out $(ICON_PNG) >/dev/null; \
	elif command -v rsvg-convert >/dev/null 2>&1; then \
		rsvg-convert --width 512 --height 512 $(ICON_SVG) --output $(ICON_PNG); \
	else \
		echo "Instale sips (macOS) ou librsvg2-bin (Linux) para gerar o ícone."; exit 1; \
	fi

test:
	go test ./...

clean:
	rm -rf $(DIST) $(ICON_PNG) $(APP).app
