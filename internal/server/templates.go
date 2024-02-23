package server

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"dario.cat/mergo"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/goccy/go-yaml"
	"github.com/sboon-gg/svctl/pkg/templates"
)

func (s *Server) Render() ([]templates.RenderOutput, error) {
	renderer, err := templates.NewFromPath(s.dotPath(TemplatesDir))
	if err != nil {
		return nil, err
	}

	values, err := s.Values()
	if err != nil {
		return nil, err
	}

	return renderer.Render(values)
}

func (s *Server) Values() (templates.Values, error) {
	var allValues templates.Values

	config, err := s.Config()
	if err != nil {
		return nil, err
	}

	for _, source := range config.Values {
		if source.File != "" {
			content, err := os.ReadFile(filepath.Join(s.dotPath(defaultValuesFile)))
			if err != nil {
				return nil, err
			}

			var values templates.Values
			err = yaml.Unmarshal(content, &values)
			if err != nil {
				return nil, err
			}

			err = mergo.Map(&allValues, values, mergo.WithOverride)
			if err != nil {
				return nil, err
			}
		}
	}

	return allValues, nil
}

func (s *Server) WriteTemplatesOutput(outputs []templates.RenderOutput) error {
	for _, out := range outputs {
		path := filepath.Join(s.Path, out.Destination)

		err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
		if err != nil {
			return err
		}

		err = os.WriteFile(path, out.Content, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

func cloneTemplates(path, repoURL, token string) error {
	var auth transport.AuthMethod
	if token != "" {
		auth = &http.BasicAuth{
			Username: "git",
			Password: token,
		}
	}

	_, err := git.PlainClone(path, false, &git.CloneOptions{
		URL:             repoURL,
		Auth:            auth,
		InsecureSkipTLS: true,
	})
	return err
}

func writeValues(path string, content []byte) error {
	commented := commentOutWholeYamlFile(string(content))

	return os.WriteFile(filepath.Join(path, defaultValuesFile), []byte(commented), 0644)
}

func commentOutWholeYamlFile(content string) string {
	reader := strings.NewReader(content)
	scanner := bufio.NewScanner(reader)

	builder := strings.Builder{}

	for scanner.Scan() {
		text := scanner.Text()

		if !strings.HasPrefix(strings.TrimSpace(text), "#") {
			if len(strings.TrimSpace(text)) > 0 {
				text = fmt.Sprintf("# %s", text)
			}
		}

		builder.WriteString(text)
		builder.WriteString("\n")
	}

	return builder.String()
}
