# Word-filter
Data structure and relevant algorithms for extremely fast prefix/fuzzy/filter string searching.

## Usage

Create a Trie with:

```Go
f := filter.New()
```

Add Keys with:

```Go
// Add can take in meta information which can be stored with the key.
// i.e. you could store any information you would like to associate with
// this particular key.
f.Add("foobar", 1)
```

Find a key with:

```Go
node, ok := t.Find("foobar")
meta := node.Meta()
// use meta with meta.(type)
```

Remove Keys with:

```Go
f.Remove("foobar")
```

Word filter with:

```Go
f.Filter("fo","*")
```

Prefix search with:

```Go
f.PrefixSearch("foo")
```

Fast test for valid prefix:
```Go
f.HasKeysWithPrefix("foo")
```

Fuzzy search with:

```Go
f.FuzzySearch("fb")
```