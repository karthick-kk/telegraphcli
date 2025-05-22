package markdown

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	telegraph "source.toby3d.me/toby3d/telegraph/v2"
	"gopkg.in/yaml.v3"
	"golang.org/x/net/html/atom"
)

// Parse parses a markdown file and returns the content as telegraph nodes
func Parse(filePath string) ([]telegraph.Node, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read markdown file: %v", err)
	}

	// Skip YAML front matter if present
	content = skipYAMLFrontMatter(content)

	// Simple markdown to telegraph nodes converter
	nodes, err := markdownToNodes(string(content))
	if err != nil {
		return nil, err
	}

	return nodes, nil
}

// skipYAMLFrontMatter removes YAML front matter from markdown content
func skipYAMLFrontMatter(content []byte) []byte {
	// Check if content starts with "---"
	if !bytes.HasPrefix(content, []byte("---\n")) {
		return content
	}

	// Find the closing "---"
	parts := bytes.SplitN(content[4:], []byte("---\n"), 2)
	if len(parts) != 2 {
		return content
	}

	// Return content after front matter
	return parts[1]
}

// markdownToNodes converts markdown content to telegraph nodes
func markdownToNodes(content string) ([]telegraph.Node, error) {
	lines := strings.Split(content, "\n")
	nodes := []telegraph.Node{}

	// Process lines
	var currentParagraph []string
	var inCodeBlock bool
	var codeContent []string
	var codeLanguage string

	flushParagraph := func() {
		if len(currentParagraph) > 0 {
			text := strings.Join(currentParagraph, " ")
			if text = strings.TrimSpace(text); text != "" {
				// Create paragraph node
				pTag, _ := telegraph.NewTag(atom.P)
				pElem := telegraph.NewNodeElement(pTag)
				
				// Create text node
				textNode := telegraph.Node{
					Text: text,
				}
				
				// Add text node to paragraph
				pElem.Children = append(pElem.Children, textNode)
				
				// Create the node with NodeElement
				node := telegraph.Node{
					Element: pElem,
				}
				
				// Add to nodes
				nodes = append(nodes, node)
			}
			currentParagraph = []string{}
		}
	}

	for _, line := range lines {
		// Handle code blocks
		if strings.HasPrefix(line, "```") {
			if inCodeBlock {
				// End of code block
				inCodeBlock = false
				code := strings.Join(codeContent, "\n")
				
				// Create pre tag
				preTag, _ := telegraph.NewTag(atom.Pre)
				preElem := telegraph.NewNodeElement(preTag)
				
				// Create code tag
				codeTag, _ := telegraph.NewTag(atom.Code)
				codeElem := telegraph.NewNodeElement(codeTag)
				
				// We can't set language attribute as Telegraph only supports href and src
				// Just keep track of the language for reference
				_ = codeLanguage
				
				// Add code text as child of code element
				codeElem.Children = append(codeElem.Children, telegraph.Node{Text: code})
				
				// Add code element as child of pre element
				preElem.Children = append(preElem.Children, telegraph.Node{Element: codeElem})
				
				// Create the node with NodeElement and add to nodes
				nodes = append(nodes, telegraph.Node{Element: preElem})
				
				codeContent = []string{}
				codeLanguage = ""
				continue
			} else {
				// Start of code block
				flushParagraph()
				inCodeBlock = true
				codeLanguage = strings.TrimSpace(strings.TrimPrefix(line, "```"))
				continue
			}
		}

		if inCodeBlock {
			codeContent = append(codeContent, line)
			continue
		}

		// Handle headers
		if strings.HasPrefix(line, "# ") {
			flushParagraph()
			text := strings.TrimSpace(strings.TrimPrefix(line, "# "))
			
			// Create h3 tag (Telegraph doesn't support h1)
			h3Tag, _ := telegraph.NewTag(atom.H3)
			h3Elem := telegraph.NewNodeElement(h3Tag)
			
			// Add text as child
			h3Elem.Children = append(h3Elem.Children, telegraph.Node{Text: text})
			
			// Create node with NodeElement and add to nodes
			nodes = append(nodes, telegraph.Node{Element: h3Elem})
			continue
		}

		if strings.HasPrefix(line, "## ") {
			flushParagraph()
			text := strings.TrimSpace(strings.TrimPrefix(line, "## "))
			
			// Create h4 tag (Telegraph doesn't support h2)
			h4Tag, _ := telegraph.NewTag(atom.H4)
			h4Elem := telegraph.NewNodeElement(h4Tag)
			
			// Add text as child
			h4Elem.Children = append(h4Elem.Children, telegraph.Node{Text: text})
			
			// Create node with NodeElement and add to nodes
			nodes = append(nodes, telegraph.Node{Element: h4Elem})
			continue
		}

		if strings.HasPrefix(line, "### ") {
			flushParagraph()
			text := strings.TrimSpace(strings.TrimPrefix(line, "### "))
			
			// Create h4 tag
			h4Tag, _ := telegraph.NewTag(atom.H4)
			h4Elem := telegraph.NewNodeElement(h4Tag)
			
			// Add text as child
			h4Elem.Children = append(h4Elem.Children, telegraph.Node{Text: text})
			
			// Create node with NodeElement and add to nodes
			nodes = append(nodes, telegraph.Node{Element: h4Elem})
			continue
		}

		// Handle lists
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			flushParagraph()
			text := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "- "), "* "))
			
			// Create li tag
			liTag, _ := telegraph.NewTag(atom.Li)
			liElem := telegraph.NewNodeElement(liTag)
			
			// Add text as child
			liElem.Children = append(liElem.Children, telegraph.Node{Text: text})
			
			// Create li node
			liNode := telegraph.Node{Element: liElem}
			
			// Check if previous node is a ul element
			var isUl bool
			if len(nodes) > 0 && nodes[len(nodes)-1].Element != nil {
				tagAtom := nodes[len(nodes)-1].Element.Tag.Atom()
				if tagAtom == atom.Ul {
					isUl = true
				}
			}
			
			if isUl {
				// Add to existing ul element
				nodes[len(nodes)-1].Element.Children = append(nodes[len(nodes)-1].Element.Children, liNode)
			} else {
				// Create new ul tag
				ulTag, _ := telegraph.NewTag(atom.Ul)
				ulElem := telegraph.NewNodeElement(ulTag)
				
				// Add li as child
				ulElem.Children = append(ulElem.Children, liNode)
				
				// Create ul node and add to nodes
				nodes = append(nodes, telegraph.Node{Element: ulElem})
			}
			continue
		}

		// Empty line - flush paragraph
		if strings.TrimSpace(line) == "" {
			flushParagraph()
			continue
		}

		// Normal paragraph text
		currentParagraph = append(currentParagraph, line)
	}

	// Flush any remaining paragraph
	flushParagraph()

	return nodes, nil
}

// ReadTitle reads the title from a markdown file's front matter
func ReadTitle(r io.Reader) (string, error) {
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		return "", err
	}
	
	content := buf.Bytes()
	
	// Check if content starts with "---" (YAML front matter)
	if !bytes.HasPrefix(content, []byte("---\n")) {
		return "", fmt.Errorf("no front matter found")
	}
	
	// Find the closing "---"
	parts := bytes.SplitN(content[4:], []byte("---\n"), 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid front matter format")
	}
	
	// Parse front matter
	var frontMatter struct {
		Title string `yaml:"title"`
	}
	
	if err := yaml.Unmarshal(parts[0], &frontMatter); err != nil {
		return "", fmt.Errorf("failed to parse front matter: %v", err)
	}
	
	if frontMatter.Title == "" {
		return "", fmt.Errorf("no title found in front matter")
	}
	
	return frontMatter.Title, nil
}
