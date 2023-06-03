.PHONY: build
build:
	CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=7 CC=arm-linux-gnueabihf-gcc CXX=arm-linux-gnueabihf-g++ go build -o bin/smart-home-server_armv7 ./cmd/server/...
	CGO_ENABLED=1 go build -o bin/smart-home-local ./cmd/server/...

.PHONY: script
script:
	CGO_ENABLED=1 go build -o bin/script ./cmd/script/...
	cd . && ./bin/script --config 'example/config.yml'

.PHONY: debug
debug: build
	./bin/smart-home-local --config 'example/config.yml'

.PHONY: deploy
deploy: build
	cd bin && (echo "smart-home-server_armv7" | pax -w | ssh pi@192.168.0.80 "cd smart-home && pax -r && mv ./smart-home-server_armv7 ./zy-smart-home") && cd ..
