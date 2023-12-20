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

//go:embed templates/actions_empty.tmpl
var actionsEmptyTmpl embed.FS

const (
	DEFAULT_TEMPLATE       string = "templates/index.tmpl"
	NO_ACTIONS_TEMPLATE    string = "templates/index_no_actions.tmpl"
	NO_GETTERS_TEMPLATE    string = "templates/index_no_getters.tmpl"
	STATE_ONLY_TEMPLATE    string = "templates/index_state_only.tmpl"
	ACTIONS_EMPTY_TEMPLATE string = "templates/actions_empty.tmpl"
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
	case ACTIONS_EMPTY_TEMPLATE:
		return actionsEmptyTmpl.ReadFile(ACTIONS_EMPTY_TEMPLATE)
	default:
		return indexTmpl.ReadFile(DEFAULT_TEMPLATE)
	}
}

func getTemplatePath(filesMap map[string]*os.File, filename string) string {
	var key = getFirstKey(filesMap)

	return strings.Replace(filesMap[key].Name(), key, filename, 1)
}

func parseTemplate(templateData string, outputPath string, values map[string]string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	tmpl := template.Must(template.New("template").Parse(templateData))

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

func createTemplate(templateType string, outputPath string, values map[string]string) error {
	data, err := getTemplate(templateType)
	if err != nil {
		return err
	}

	err = parseTemplate(string(data), outputPath, values)
	if err != nil {
		return err
	}

	return nil
}
