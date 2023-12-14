<p align="center">
  <h3 align="center">Vuex2Pinia</h3>
  <p align="center">A migration tool to translate vuex code into pinia code<p>
</p>

[![Go Report Card](https://goreportcard.com/badge/github.com/fdbiondi/vuex-pinia-migration-tool)](https://goreportcard.com/report/github.com/basebandit/gocash)  [![GitHub license](https://img.shields.io/github/license/fdbiondi/vuex-pinia-migration-tool)](https://github.com/fdbiondi/vuex-pinia-migration-tool/blob/main/LICENSE)

## Install

```
git clone https://github.com/fdbiondi/vuex-pinia-migration-tool.git
make build
make install
```

## Requirements

```
# Required Vuex directory structure
src
└── store
	├── module1
	|		├── actions.ts
	|		├── getters.ts
	|		├── mutations.ts
	|		└── state.ts
	└── module2
			├── actions.ts
			├── getters.ts
			├── mutations.ts
			└── state.ts
```


## Usage

```bash
	$ vuex-to-pinia migrate <from> <to>
	$ vuex-to-pinia <options>
```

## Contributing

Contributions, issues and feature requests are welcome! 👍 <br> Feel free to

check [open issues](https://github.com/fdbiondi/vuex-pinia-migration-tool/issues).

## Quick Start

```bash
	git clone https://github.com/fdbiondi/vuex-pinia-migration-tool.git
	cd vuex-pinia-migration-tool
	go get -d ./...
	go run cmd/main.go
```

## License

(c) 2023 [MIT Licence](https://opensource.org/licenses/MIT)
