package properties

import (
	"bufio"
	"errors"
	"fmt"
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

type NodeNotFoundError struct{}

func (nnf *NodeNotFoundError) Error() string {
	return "Node not found"
}

type NoValueError struct {
	key string
}

func (nv *NoValueError) Error() string {
	return "NoValueError: node '" + nv.key + "' has no value"
}

type ParseError struct {
	msg string
}

func (pe *ParseError) Error() string {
	return "ParseError: " + pe.msg
}

var (
	COMMENT_REPLACE_RE = regexp.MustCompile(`\s*#.*$`)
	ASSIGNMENT_RE      = regexp.MustCompile(`^\s*([\w\.]+)\s*=\s*(.*)$`)
	SECTION_RE         = regexp.MustCompile(`^\s*\[([\w\.]+)\]\s*$`)
	NodeNotFound       = &NodeNotFoundError{}
	TrueValues         = []string{"true", "yes", "1"}
	FalseValues        = []string{"false", "no", "0"}
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
		return nil, NodeNotFound
	}
	return childNode.findNode(key)
}

type Properties struct {
	filename string
	root     *propertiesNode
}

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

func (p *Properties) GetBool(key string) (bool, error) {
	value, err := p.GetString(key)
	if err != nil {
		return false, err
	}
	value = strings.ToLower(value)
	for _, tv := range TrueValues {
		if value == tv {
			return true, nil
		}
	}
	for _, fv := range FalseValues {
		if value == fv {
			return false, nil
		}
	}
	return false, errors.New(fmt.Sprintf("invalid boolean value \"%s\"", value))
}

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

func (p *Properties) KeyExists(key string) bool {
	_, err := p.root.findNode(key)
	return err != NodeNotFound
}

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

func (p *Properties) parseFile() error {
	f, err := os.Open(p.filename)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	currentSection := ""
	for scanner.Scan() {
		text := scanner.Text()
		line := COMMENT_REPLACE_RE.ReplaceAllString(text, "")
		line = strings.Trim(line, " \t\n")
		if line == "" {
			continue
		}

		match := SECTION_RE.FindStringSubmatch(line)
		if len(match) > 0 {
			currentSection = match[1] + "."
			continue
		}

		match = ASSIGNMENT_RE.FindStringSubmatch(line)
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

func Load(filename string) (*Properties, error) {
	p := &Properties{
		filename,
		newNode("", "", nil),
	}
	err := p.parseFile()
	if err != nil {
		return nil, err
	}
	return p, nil
}
