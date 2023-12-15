build:
	go build -o bin/vuex-to-pinia cmd/main.go

clean:
	go clean
	rm bin/vuex-to-pinia

install:
	cp bin/vuex-to-pinia ~/.local/bin/

uninstall:
	rm ~/.local/bin/vuex-to-pinia

