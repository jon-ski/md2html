package main

import (
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"golang.org/x/net/html"
	"gopkg.in/yaml.v3"
)

type Config struct {
	InlineStyles map[string]string `yaml:"inline_styles"`
}

func loadConfig() Config {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Printf("failed to get home dir: %v", err)
		return Config{}
	}

	configPath := filepath.Join(home, ".config", "md2html", "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return Config{} // No config, just fallback
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Printf("failed to parse config: %v", err)
		return Config{}
	}
	return cfg
}

func stripFrontmatter(input []byte) []byte {
	lines := bytes.Split(input, []byte("\n"))
	if len(lines) == 0 || !bytes.Equal(bytes.TrimSpace(lines[0]), []byte("---")) {
		return input
	}

	for i := 1; i < len(lines); i++ {
		if bytes.Equal(bytes.TrimSpace(lines[i]), []byte("---")) {
			return bytes.Join(lines[i+1:], []byte("\n"))
		}
	}
	return input // no closing ---
}

func applyInlineStyles(input string, styles map[string]string) string {
	node, err := html.Parse(strings.NewReader(input))
	if err != nil {
		return input // fallback on failure
	}

	applyStylesRecursive(node, styles)

	var buf bytes.Buffer
	if err := html.Render(&buf, node); err != nil {
		return input // fallback
	}

	return buf.String()
}

func applyStylesRecursive(n *html.Node, styles map[string]string) {
	if n.Type == html.ElementNode {
		if style, ok := styles[n.Data]; ok {
			hasStyle := false
			for i, attr := range n.Attr {
				if attr.Key == "style" {
					hasStyle = true
					n.Attr[i].Val = attr.Val + "; " + style
					break
				}
			}
			if !hasStyle {
				n.Attr = append(n.Attr, html.Attribute{Key: "style", Val: style})
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		applyStylesRecursive(c, styles)
	}
}

func main() {
	md := goldmark.New(goldmark.WithExtensions(
		extension.GFM,
		extension.Footnote,
	))

	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("failed to read stdin: %v", err)
	}

	input = stripFrontmatter(input)

	var buf strings.Builder
	if err := md.Convert(input, &buf); err != nil {
		log.Fatalf("failed to convert: %v", err)
	}

	cfg := loadConfig()
	output := applyInlineStyles(buf.String(), cfg.InlineStyles)

	_, err = os.Stdout.Write([]byte(output))
	if err != nil {
		log.Fatalf("failed to write to stdout: %v", err)
	}
}
