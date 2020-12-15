package filter

import (
	"sort"
	"sync"
)

type Node struct {
	val       rune
	path      string
	term      bool
	end       bool
	depth     int
	meta      interface{}
	mask      uint64
	parent    *Node
	children  map[rune]*Node
	termCount int
}

type Trie struct {
	mu   sync.Mutex
	root *Node
	size int
}

type ByKeys []string

func (a ByKeys) Len() int           { return len(a) }
func (a ByKeys) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKeys) Less(i, j int) bool { return len(a[i]) < len(a[j]) }

const nul = 0x0

func New() *Trie {
	return &Trie{
		root: &Node{children: make(map[rune]*Node), depth: 0},
		size: 0,
	}
}

func (t *Trie) Root() *Node {
	return t.root
}

func (t *Trie) Add(key string, meta interface{}) *Node {
	t.mu.Lock()

	t.size++
	runes := []rune(key)
	bitmask := maskruneslice(runes)
	node := t.root
	node.mask |= bitmask
	node.termCount++
	for i := range runes {
		r := runes[i]
		bitmask = maskruneslice(runes[i:])
		if n, ok := node.children[r]; ok {
			node = n
			node.mask |= bitmask
		} else {
			node = node.NewChild(r, "", bitmask, nil, false)
		}
		node.termCount++
		if i == len(runes)-1 {
			node.end = true
		}
	}
	node = node.NewChild(nul, key, 0, meta, true)
	t.mu.Unlock()
	return node
}

func (t *Trie) Find(key string) (*Node, bool) {
	node := findNode(t.Root(), []rune(key))
	if node == nil {
		return nil, false
	}

	node, ok := node.Children()[nul]
	if !ok || !node.term {
		return nil, false
	}

	return node, true
}

func (t *Trie) HasKeysWithPrefix(key string) bool {
	node := findNode(t.Root(), []rune(key))
	return node != nil
}

func (t *Trie) Remove(key string) {
	var (
		i    int
		rs   = []rune(key)
		node = findNode(t.Root(), []rune(key))
	)
	t.mu.Lock()

	t.size--
	for n := node.Parent(); n != nil; n = n.Parent() {
		i++
		if len(n.Children()) > 1 {
			r := rs[len(rs)-i]
			n.RemoveChild(r)
			break
		}
	}
	t.mu.Unlock()
}

func (t *Trie) Keys() []string {
	if t.size == 0 {
		return []string{}
	}

	return t.PrefixSearch("")
}

func (t *Trie) FuzzySearch(pre string) []string {
	keys := fuzzycollect(t.Root(), []rune(pre))
	sort.Sort(ByKeys(keys))
	return keys
}

func (t *Trie) PrefixSearch(pre string) []string {
	node := findNode(t.Root(), []rune(pre))
	if node == nil {
		return nil
	}

	return collect(node)
}

func (parent *Node) NewChild(val rune, path string, bitmask uint64, meta interface{}, term bool) *Node {
	node := &Node{
		val:      val,
		path:     path,
		mask:     bitmask,
		term:     term,
		meta:     meta,
		parent:   parent,
		children: make(map[rune]*Node),
		depth:    parent.depth + 1,
	}
	parent.children[node.val] = node
	parent.mask |= bitmask
	return node
}

func (n *Node) RemoveChild(r rune) {
	delete(n.children, r)
	for nd := n.parent; nd != nil; nd = nd.parent {
		nd.mask ^= nd.mask
		nd.mask |= uint64(1) << uint64(nd.val-'a')
		for _, c := range nd.children {
			nd.mask |= c.mask
		}
	}
}

func (n Node) Parent() *Node {
	return n.parent
}

func (n Node) Meta() interface{} {
	return n.meta
}

func (n Node) Children() map[rune]*Node {
	return n.children
}

func (n Node) Terminating() bool {
	return n.term
}

func (n Node) Val() rune {
	return n.val
}

func (n Node) Depth() int {
	return n.depth
}

func (n Node) Mask() uint64 {
	return n.mask
}

func findNode(node *Node, runes []rune) *Node {
	if node == nil {
		return nil
	}

	if len(runes) == 0 {
		return node
	}

	n, ok := node.Children()[runes[0]]
	if !ok {
		return nil
	}

	var nrunes []rune
	if len(runes) > 1 {
		nrunes = runes[1:]
	} else {
		nrunes = runes[0:0]
	}

	return findNode(n, nrunes)
}

func maskruneslice(rs []rune) uint64 {
	var m uint64
	for _, r := range rs {
		m |= uint64(1) << uint64(r-'a')
	}
	return m
}

func collect(node *Node) []string {
	var (
		n *Node
		i int
	)
	var keys []string
	nodes := make([]*Node, 1, len(node.children))
	nodes[0] = node
	for l := len(nodes); l != 0; l = len(nodes) {
		i = l - 1
		n = nodes[i]
		nodes = nodes[:i]
		for _, c := range n.children {
			nodes = append(nodes, c)
		}
		if n.term {
			word := n.path
			keys = append(keys, word)
		}
	}
	return keys
}

type potentialSubtree struct {
	idx  int
	node *Node
}

func fuzzycollect(node *Node, partial []rune) []string {
	if len(partial) == 0 {
		return collect(node)
	}

	var (
		m    uint64
		i    int
		p    potentialSubtree
		keys []string
	)

	potential := []potentialSubtree{potentialSubtree{node: node, idx: 0}}
	for l := len(potential); l > 0; l = len(potential) {
		i = l - 1
		p = potential[i]
		potential = potential[:i]
		m = maskruneslice(partial[p.idx:])
		if (p.node.mask & m) != m {
			continue
		}

		if p.node.val == partial[p.idx] {
			p.idx++
			if p.idx == len(partial) {
				keys = append(keys, collect(p.node)...)
				continue
			}
		}

		for _, c := range p.node.children {
			potential = append(potential, potentialSubtree{node: c, idx: p.idx})
		}
	}
	return keys
}

func (t *Trie) Filter(text, replace string) (result string, hit bool) {
	chars := []rune(text)
	if t.root == nil {
		return
	}
	var left []rune
	node := t.root
	var start int
	for k := 0; k < len(chars); k++ {
		var macthEnd bool
		for j := k; j < len(chars); j++ {
			if ret, ok := node.children[chars[j]]; !ok {
				node = t.root
				if macthEnd { //之前已匹配
					for n := start; n < j; n++ {
						left = append(left, ([]rune(replace))...)
					}
					start = j
					k = j - 1
					break
				}
				left = append(left, chars[start:k+1]...)
				start = k + 1
				break
			} else {
				node = ret
				if ret.end {
					macthEnd = true
					hit = true
					if j == len(chars)-1 {
						for n := start; n < j+1; n++ {
							left = append(left, ([]rune(replace))...)
						}
						start = j + 1
						k = j
					}
				}
			}
		}
	}
	result = string(left)
	return
}
