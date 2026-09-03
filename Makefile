APP := Highway
APP_ID := com.herrmann.highway
PACKAGE := highway
VERSION ?= 0.1.0
ARCH ?= $(shell go env GOARCH)
DEB_REVISION ?= 1
DEB_VERSION := $(VERSION)-$(DEB_REVISION)
DIST := dist
ICON_SVG := assets/highway.svg
ICON_PNG := $(DIST)/highway.png
FYNE := go run fyne.io/tools/cmd/fyne@latest
LDFLAGS := -X main.appVersion=$(VERSION)
DEB_ROOT := $(DIST)/deb/$(PACKAGE)_$(DEB_VERSION)_amd64
DMG_STAGE := $(DIST)/dmg
DMG := $(DIST)/$(APP)-$(VERSION)-macos-$(ARCH).dmg

.PHONY: macos macos-dmg macos-service linux deb icon clean test

macos: icon
	mkdir -p $(DIST)
	rm -rf $(DIST)/$(APP).app $(APP).app
	go build -ldflags "$(LDFLAGS)" -o $(DIST)/$(APP) .
	$(FYNE) package -os darwin -name $(APP) -appID $(APP_ID) -icon $(ICON_PNG) --executable $(DIST)/$(APP)
	rm -f $(DIST)/$(APP)
	mv $(APP).app $(DIST)/$(APP).app

macos-dmg: macos
	rm -rf $(DMG_STAGE) $(DMG)
	mkdir -p $(DMG_STAGE)
	cp -R $(DIST)/$(APP).app $(DMG_STAGE)/$(APP).app
	ln -s /Applications $(DMG_STAGE)/Applications
	hdiutil create -volname $(APP) -srcfolder $(DMG_STAGE) -ov -format UDZO $(DMG)
	hdiutil verify $(DMG)
	rm -rf $(DMG_STAGE)

macos-service:
	zsh packaging/macos/install-service.sh

linux: icon
	@if [ "$$(go env GOOS)" != "linux" ] || [ "$$(go env GOARCH)" != "amd64" ]; then \
		echo "Execute 'make linux' em uma maquina Linux x86_64."; exit 1; \
	fi
	mkdir -p $(DIST)
	go build -ldflags "$(LDFLAGS)" -o $(DIST)/highway .
	cp packaging/highway.desktop $(DIST)/highway.desktop
	cp $(ICON_PNG) $(DIST)/highway.png
	tar -czf $(DIST)/highway-linux-amd64.tar.gz -C $(DIST) highway highway.desktop highway.png

deb: icon
	@if [ "$$(go env GOOS)" != "linux" ] || [ "$$(go env GOARCH)" != "amd64" ]; then \
		echo "Execute 'make deb' em uma maquina Linux x86_64."; exit 1; \
	fi
	rm -rf $(DEB_ROOT)
	install -d $(DEB_ROOT)/usr/bin
	go build -ldflags "$(LDFLAGS)" -o $(DEB_ROOT)/usr/bin/$(PACKAGE) .
	install -D -m 0644 packaging/highway.desktop $(DEB_ROOT)/usr/share/applications/$(PACKAGE).desktop
	install -D -m 0644 $(ICON_PNG) $(DEB_ROOT)/usr/share/icons/hicolor/512x512/apps/$(PACKAGE).png
	install -D -m 0644 packaging/debian/control.in $(DEB_ROOT)/DEBIAN/control
	sed -i 's/@VERSION@/$(DEB_VERSION)/' $(DEB_ROOT)/DEBIAN/control
	install -D -m 0644 packaging/debian/copyright $(DEB_ROOT)/usr/share/doc/$(PACKAGE)/copyright
	install -D -m 0644 packaging/debian/changelog $(DEB_ROOT)/usr/share/doc/$(PACKAGE)/changelog.Debian
	gzip -n -f $(DEB_ROOT)/usr/share/doc/$(PACKAGE)/changelog.Debian
	dpkg-deb --build --root-owner-group $(DEB_ROOT) $(DIST)/$(PACKAGE)_$(DEB_VERSION)_amd64.deb

icon:
	mkdir -p $(DIST)
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
