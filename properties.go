package properties

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type propertiesNode struct {
	root  *propertiesNode
	key   string
	value string
	tree  map[string]*propertiesNode
}

// NodeNotFoundError represents an error
type NodeNotFoundError struct{}

func (nnf *NodeNotFoundError) Error() string {
	return "node not found"
}

// NoValueError represents an error of an empty-value node
type NoValueError struct {
	key string
}

func (nv *NoValueError) Error() string {
	return "node '" + nv.key + "' has no value"
}

// ParseError represents a parsing error
type ParseError struct {
	msg string
}

func (pe *ParseError) Error() string {
	return "parse error: " + pe.msg
}

var (
	commentReplaceExpr = regexp.MustCompile(`\s*#.*$`)
	assignmentExpr     = regexp.MustCompile(`^\s*([\w\.]+)\s*=\s*(.*)$`)
	sectionExpr        = regexp.MustCompile(`^\s*\[([\w\.]+)\]\s*$`)
	nodeNotFound       = &NodeNotFoundError{}
	trueValues         = []string{"true", "yes", "1"}
	falseValues        = []string{"false", "no", "0"}
)

func newNode(key string, value string, root *propertiesNode) *propertiesNode {
	n := &propertiesNode{
		root,
		key,
		value,
		make(map[string]*propertiesNode),
	}
	if root == nil {
		n.root = n
	}
	return n
}

func (n *propertiesNode) put(key, value string) {
	partialKey := key
	if n.key != "" {
		if key == n.key {
			n.value = value
			return
		}
		if !strings.HasPrefix(key, n.key+".") {
			n.root.put(key, value)
			return
		}
		partialKey = strings.Replace(key, n.key+".", "", 1)
	}

	tokens := strings.Split(partialKey, ".")

	firstToken := tokens[0]
	childNode, ok := n.tree[firstToken]

	if !ok {
		prefix := n.key
		if prefix != "" {
			prefix += "."
		}
		n.tree[firstToken] = newNode(prefix+firstToken, "", n.root)
		childNode, _ = n.tree[firstToken]
	}

	childNode.put(key, value)
}

func (n *propertiesNode) findNode(key string) (*propertiesNode, error) {
	partialKey := key
	if n.key != "" {
		if key == n.key {
			return n, nil
		}
		if !strings.HasPrefix(key, n.key+".") {
			return n.root.findNode(key)
		}
		partialKey = strings.Replace(key, n.key+".", "", 1)
	}
	tokens := strings.Split(partialKey, ".")

	firstToken := tokens[0]
	childNode, ok := n.tree[firstToken]

	if !ok {
		return nil, nodeNotFound
	}
	return childNode.findNode(key)
}

// Properties is the main module struct
type Properties struct {
	rd   io.Reader
	root *propertiesNode
}

// GetString returns a string by a given key
func (p *Properties) GetString(key string) (string, error) {
	node, err := p.root.findNode(key)
	if err != nil {
		return "", err
	}
	if node.value == "" {
		return "", &NoValueError{key}
	}
	return node.value, nil
}

// GetInt returns an int by a given key
func (p *Properties) GetInt(key string) (int, error) {
	value, err := p.GetString(key)
	if err != nil {
		return 0, err
	}
	int64Value, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, err
	}
	return int(int64Value), nil
}

// GetBool returns a boolean by a given key
func (p *Properties) GetBool(key string) (bool, error) {
	value, err := p.GetString(key)
	if err != nil {
		return false, err
	}
	value = strings.ToLower(value)
	for _, tv := range trueValues {
		if value == tv {
			return true, nil
		}
	}
	for _, fv := range falseValues {
		if value == fv {
			return false, nil
		}
	}
	return false, errors.New(fmt.Sprintf("invalid boolean value \"%s\"", value))
}

// GetFloat returns a float value by a given key
func (p *Properties) GetFloat(key string) (float64, error) {
	value, err := p.GetString(key)
	if err != nil {
		return 0.0, err
	}
	f64Value, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0.0, err
	}
	return f64Value, nil
}

// KeyExists checks if a key exists
func (p *Properties) KeyExists(key string) bool {
	_, err := p.root.findNode(key)
	return err != nodeNotFound
}

// Subkeys returns a list of subkeys of a given key
func (p *Properties) Subkeys(key string) ([]string, error) {
	subkeys := make([]string, 0)

	node, err := p.root.findNode(key)
	if err != nil {
		return subkeys, err
	}

	for key := range node.tree {
		subkeys = append(subkeys, key)
	}
	return subkeys, nil
}

func (p *Properties) parse() error {

	scanner := bufio.NewScanner(p.rd)
	currentSection := ""
	for scanner.Scan() {
		text := scanner.Text()
		line := commentReplaceExpr.ReplaceAllString(text, "")
		line = strings.Trim(line, " \t\n")
		if line == "" {
			continue
		}

		match := sectionExpr.FindStringSubmatch(line)
		if len(match) > 0 {
			currentSection = match[1] + "."
			continue
		}

		match = assignmentExpr.FindStringSubmatch(line)
		if len(match) > 0 {
			key := currentSection + match[1]
			value := match[2]
			p.root.put(key, value)
			continue
		}

		return &ParseError{"invalid line: " + text}
	}
	return nil
}

// Load loads a file with a given filename, parses it and returns
// a newly configured Properties object
func Load(filename string) (*Properties, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	p := &Properties{f, newNode("", "", nil)}

	err = p.parse()
	if err != nil {
		return nil, err
	}
	return p, nil
}
