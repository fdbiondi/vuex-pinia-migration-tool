package parser

import (
	"bytes"
	"embed"
	"errors"
	"os"
	"strings"
	"text/template"
)

//go:embed templates/index.tmpl
var indexTmpl embed.FS

//go:embed templates/index_no_actions.tmpl
var indexNoActionsTmpl embed.FS

//go:embed templates/index_no_getters.tmpl
var indexNoGettersTmpl embed.FS

//go:embed templates/index_state_only.tmpl
var indexStateOnlyTmpl embed.FS

const (
	DEFAULT_TEMPLATE    string = "templates/index.tmpl"
	NO_ACTIONS_TEMPLATE string = "templates/index_no_actions.tmpl"
	NO_GETTERS_TEMPLATE string = "templates/index_no_getters.tmpl"
	STATE_ONLY_TEMPLATE string = "templates/index_state_only.tmpl"
)

func getFirstKey[K comparable, V any](m map[K]V) K {
	for k := range m {
		return k
	}

	return *new(K)
}

func getTemplate(template string) ([]byte, error) {
	switch template {
	case NO_ACTIONS_TEMPLATE:
		return indexNoActionsTmpl.ReadFile(NO_ACTIONS_TEMPLATE)
	case NO_GETTERS_TEMPLATE:
		return indexNoGettersTmpl.ReadFile(NO_GETTERS_TEMPLATE)
	case STATE_ONLY_TEMPLATE:
		return indexStateOnlyTmpl.ReadFile(STATE_ONLY_TEMPLATE)
	default:
		return indexTmpl.ReadFile(DEFAULT_TEMPLATE)
	}
}

func createIndexTemplate(filesMap map[string]*os.File, template string) (string, error) {
	key := getFirstKey(filesMap)

	var storeFilename = strings.Replace(filesMap[key].Name(), key, "index", 1)

	pathSlice := strings.Split(storeFilename, "/")
	storeName := pathSlice[len(pathSlice)-2]

	data, err := getTemplate(template)
	if err != nil {
		return "", err
	}

	var values = map[string]string{
		"storeName":          storeName,
		"storeNameTitleCase": kebabToCamelCase(storeName, true),
	}

	err = parseTemplate(string(data), storeFilename, values)
	if err != nil {
		return "", err
	}

	return storeName, nil
}

func parseTemplate(templateStr string, outputPath string, values map[string]string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	tmpl := template.Must(template.New("template").Parse(templateStr))

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, values); err != nil {
		return err
	}

	err = os.WriteFile(outputPath, buf.Bytes(), 0644)
	if err != nil {
		return errors.New("error wrinting file -> " + err.Error())
	}

	return nil
}
