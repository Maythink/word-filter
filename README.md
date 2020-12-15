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
f.Add("大法", 1)
f.Add("大法师", 1)
```

Find a key with:

```Go
node, ok := t.Find("大法")
meta := node.Meta()
// use meta with meta.(type)
```

Remove Keys with:

```Go
f.Remove("大法师")
```

Word filter with:

```Go
f.Filter("大法师","*")
```

Prefix search with:

```Go
f.PrefixSearch("大")
```

Fast test for valid prefix:
```Go
f.HasKeysWithPrefix("大")
```

Fuzzy search with:

```Go
f.FuzzySearch("大")
```