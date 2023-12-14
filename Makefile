build:
	go build -o bin/vuex-to-pinia cmd/main.go

install:
	mv bin/vuex-to-pinia ~/.local/bin/

