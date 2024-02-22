package templates

import (
	"bytes"
	"io/fs"
	"os"
	"text/template"

	"dario.cat/mergo"
	"gopkg.in/yaml.v3"
)

const (
	configFileName = "config.yaml"
)

type Values map[string]any

type Data struct {
	Values Values
}

type Template struct {
	Source      string `yaml:"src"`
	Destination string `yaml:"dest"`
	Reloadable  bool   `yaml:"reloadable"`
}

type RenderOutput struct {
	Template
	Content []byte
}

type Config struct {
	Templates []Template `yaml:"templates"`
	Defaults  []string   `yaml:"defaults"`
}

func ReadConfig(dir fs.FS) (*Config, error) {
	content, err := fs.ReadFile(dir, configFileName)
	if err != nil {
		return nil, err
	}

	var conf Config
	err = yaml.Unmarshal(content, &conf)
	if err != nil {
		return nil, err
	}

	return &conf, nil
}

type Renderer struct {
	config *Config
	files  fs.FS
}

func New(config *Config, files fs.FS) *Renderer {
	return &Renderer{
		config: config,
		files:  files,
	}
}

func NewFromPath(path string) (*Renderer, error) {
	dir := os.DirFS(path)
	return NewFromFS(dir)
}

func NewFromFS(files fs.FS) (*Renderer, error) {
	conf, err := ReadConfig(files)
	if err != nil {
		return nil, err
	}

	return New(conf, files), nil
}

func (t *Renderer) template(name, tplContent string) (*template.Template, error) {
	return template.New(name).Funcs(FuncMap()).Parse(tplContent)
}

func (t *Renderer) prepData(values Values) (*Data, error) {
	data := Data{
		Values: Values{},
	}

	defaults, err := t.Defaults()
	if err != nil {
		return nil, err
	}

	err = mergo.Map(&data.Values, defaults)
	if err != nil {
		return nil, err
	}

	err = mergo.Map(&data.Values, values, mergo.WithOverride)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (t *Renderer) render(name, template string, values Values) ([]byte, error) {
	tmpl, err := t.template(name, template)
	if err != nil {
		return nil, err
	}

	data, err := t.prepData(values)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (t *Renderer) Render(values Values) ([]RenderOutput, error) {
	rendered := make([]RenderOutput, len(t.config.Templates))

	for i, tmplSpec := range t.config.Templates {
		content, err := fs.ReadFile(t.files, tmplSpec.Source)
		if err != nil {
			return nil, err
		}

		out, err := t.render(tmplSpec.Source, string(content), values)
		if err != nil {
			return nil, err
		}

		rendered[i] = RenderOutput{
			Template: tmplSpec,
			Content:  out,
		}
	}

	// Second run to allow for values interpolation in values files
	for i, out := range rendered {
		content, err := t.render(out.Source, string(out.Content), values)
		if err != nil {
			return rendered, err
		}

		rendered[i].Content = content
	}

	return rendered, nil
}

func (t *Renderer) DefaultsContent() ([]byte, error) {
	var content bytes.Buffer
	for _, path := range t.config.Defaults {
		defaultsContent, err := fs.ReadFile(t.files, path)
		if err != nil {
			return nil, err
		}

		content.Write(defaultsContent)
		content.WriteString("\n")
	}

	return content.Bytes(), nil
}

func (t *Renderer) Defaults() (Values, error) {
	defaultsContent, err := t.DefaultsContent()
	if err != nil {
		return nil, err
	}

	defaults := make(Values)
	err = yaml.Unmarshal(defaultsContent, &defaults)
	if err != nil {
		return nil, err
	}

	return defaults, nil
}
