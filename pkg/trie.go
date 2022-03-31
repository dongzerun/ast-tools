package pkg

import (
	"fmt"
	"strings"
)

// Trie ...
type Trie struct {
	root     bool
	name     string
	ignore   bool
	children map[string]*Trie
}

// Search ...
func (t *Trie) Search(path string) *Trie {
	if t == nil {
		return nil
	}
	return t.children[path]
}

// IsIgnore ...
func (t *Trie) IsIgnore() bool {
	if t == nil {
		return false
	}
	return t.ignore
}

// Print ...
func (t *Trie) Print(layer int) {
	fmt.Printf("layer: %d name: %s ignore: %v\n", layer, t.name, t.ignore)
	for path := range t.children {
		t.children[path].Print(layer + 1)
		fmt.Println()
	}
}

// BuildIgnoreFieldsTrie ...
func BuildIgnoreFieldsTrie(fields []string) *Trie {
	root := &Trie{
		root:     true,
		children: map[string]*Trie{},
	}
	for i := range fields {
		node := root

		paths := strings.Split(fields[i], ".")
		for j := range paths {
			_, exists := node.children[paths[j]]
			if !exists {
				tmp := &Trie{
					name:     paths[j],
					children: map[string]*Trie{},
				}
				node.children[paths[j]] = tmp
			}

			node = node.children[paths[j]]
			if j == len(paths)-1 {
				node.ignore = true
			}
		}
	}
	return root
}
