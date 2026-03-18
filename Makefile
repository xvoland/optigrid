APP_NAME=optigrid
SRC=main.go
BUILD_DIR=build
LDFLAGS=-ldflags="-s -w"
# For Windows, add a flag to hide the console.
WIN_LDFLAGS=-ldflags="-s -w -H windowsgui"

.PHONY: all clean windows linux darwin

all: clean windows linux darwin

windows:
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build $(WIN_LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)_windows_amd64.exe $(SRC)

linux:
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)_linux_amd64 $(SRC)

darwin:
	@echo "Building for macOS Intel..."
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)_macos_amd64 $(SRC)
	@echo "Building for macOS Apple Silicon..."
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)_macos_arm64 $(SRC)

# --- UNIVERSAL BINARY ---
universal: darwin
	@echo "Creating universal binary..."
	mkdir -p $(BUILD_DIR)
	lipo -create \
		$(BUILD_DIR)/$(BIN_NAME)_amd64 \
		$(BUILD_DIR)/$(BIN_NAME)_arm64 \
		-output $(BUILD_DIR)/$(BIN_NAME)

# --- ICON (.icns from PNG) ---
icon:
	@echo "Generating .icns icon..."
	rm -rf $(ICONSET)
	mkdir $(ICONSET)

	sips -z 16 16     $(ICON_PNG) --out $(ICONSET)/icon_16x16.png
	sips -z 32 32     $(ICON_PNG) --out $(ICONSET)/icon_16x16@2x.png
	sips -z 32 32     $(ICON_PNG) --out $(ICONSET)/icon_32x32.png
	sips -z 64 64     $(ICON_PNG) --out $(ICONSET)/icon_32x32@2x.png
	sips -z 128 128   $(ICON_PNG) --out $(ICONSET)/icon_128x128.png
	sips -z 256 256   $(ICON_PNG) --out $(ICONSET)/icon_128x128@2x.png
	sips -z 256 256   $(ICON_PNG) --out $(ICONSET)/icon_256x256.png
	sips -z 512 512   $(ICON_PNG) --out $(ICONSET)/icon_256x256@2x.png
	sips -z 512 512   $(ICON_PNG) --out $(ICONSET)/icon_512x512.png
	cp $(ICON_PNG) $(ICONSET)/icon_512x512@2x.png

	iconutil -c icns $(ICONSET)
	mv icon.icns $(ICON_ICNS)

# --- APP BUNDLE ---
app: universal icon
	@echo "Creating .app bundle..."
	mkdir -p $(BUILD_DIR)/$(APP_NAME).app/Contents/MacOS
	mkdir -p $(BUILD_DIR)/$(APP_NAME).app/Contents/Resources

	cp $(BUILD_DIR)/$(BIN_NAME) $(BUILD_DIR)/$(APP_NAME).app/Contents/MacOS/$(BIN_NAME)
	cp $(ICON_ICNS) $(BUILD_DIR)/$(APP_NAME).app/Contents/Resources/

	@echo "Generating Info.plist..."
	echo '<?xml version="1.0" encoding="UTF-8"?>' > $(BUILD_DIR)/$(APP_NAME).app/Contents/Info.plist
	echo '<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"' >> $(BUILD_DIR)/$(APP_NAME).app/Contents/Info.plist
	echo ' "http://www.apple.com/DTDs/PropertyList-1.0.dtd">' >> $(BUILD_DIR)/$(APP_NAME).app/Contents/Info.plist
	echo '<plist version="1.0"><dict>' >> $(BUILD_DIR)/$(APP_NAME).app/Contents/Info.plist
	echo '<key>CFBundleName</key><string>$(APP_NAME)</string>' >> $(BUILD_DIR)/$(APP_NAME).app/Contents/Info.plist
	echo '<key>CFBundleExecutable</key><string>$(BIN_NAME)</string>' >> $(BUILD_DIR)/$(APP_NAME).app/Contents/Info.plist
	echo '<key>CFBundleIdentifier</key><string>com.example.$(BIN_NAME)</string>' >> $(BUILD_DIR)/$(APP_NAME).app/Contents/Info.plist
	echo '<key>CFBundleVersion</key><string>1.0</string>' >> $(BUILD_DIR)/$(APP_NAME).app/Contents/Info.plist
	echo '<key>CFBundlePackageType</key><string>APPL</string>' >> $(BUILD_DIR)/$(APP_NAME).app/Contents/Info.plist
	echo '<key>CFBundleIconFile</key><string>icon</string>' >> $(BUILD_DIR)/$(APP_NAME).app/Contents/Info.plist
	echo '<key>LSMinimumSystemVersion</key><string>11.0</string>' >> $(BUILD_DIR)/$(APP_NAME).app/Contents/Info.plist
	echo '</dict></plist>' >> $(BUILD_DIR)/$(APP_NAME).app/Contents/Info.plist

# --- PRETTY DMG ---
dmg: app
	@echo "Creating pretty DMG..."
	create-dmg \
	  --volname "$(APP_NAME)" \
	  --window-pos 200 120 \
	  --window-size 800 400 \
	  --icon-size 100 \
	  --icon "$(APP_NAME).app" 200 190 \
	  --hide-extension "$(APP_NAME).app" \
	  --app-drop-link 600 185 \
	  --background $(DMG_BG) \
	  $(BUILD_DIR)/$(APP_NAME).dmg \
	  $(BUILD_DIR)/$(APP_NAME).app

clean:
	@echo "Cleaning up..."
	rm -rf $(BUILD_DIR)
