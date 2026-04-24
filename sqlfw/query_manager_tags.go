package sqlfw

import "strings"

type parsedTag struct {
	kv map[string]string
}

func ParseTag(tag string) parsedTag {
	t := parsedTag{kv: make(map[string]string)}
	kvs := strings.Split(tag, ";")
	for _, kv := range kvs {
		key, value, ok := strings.Cut(kv, "=")
		if !ok {
			value = ""
		}
		t.kv[key] = value
	}
	return t
}
