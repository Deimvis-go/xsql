package sqlfw

import (
	"errors"
	"path/filepath"
	"strings"
)

type QueryPathResolver func(QueryFieldMeta) (string, error)
type QueryNameResolver func(QueryFieldMeta) (string, error)

var (
	QueryPathFromFieldTag QueryPathResolver = func(m QueryFieldMeta) (string, error) {
		// TODO: support tag path=!infer(field_name)
		if path, ok := m.FieldTag().kv["path"]; ok {
			return path, nil
		}
		return "", errors.New("no 'path' key in tag")
	}
	QueryPathFromFieldName QueryPathResolver = func(m QueryFieldMeta) (string, error) {
		return toSnake(m.FieldName()) + ".sql", nil
	}
)

var (
	QueryNameFromFieldTag QueryNameResolver = func(m QueryFieldMeta) (string, error) {
		// TODO: support tag name=!infer(field_nmae)
		if path, ok := m.FieldTag().kv["name"]; ok {
			return path, nil
		}
		return "", errors.New("no 'name' key in tag")
	}
	QueryNameFromFieldName QueryNameResolver = func(m QueryFieldMeta) (string, error) {
		return toSnake(m.FieldName()), nil
	}
	QueryNameFromFileBasename QueryNameResolver = func(m QueryFieldMeta) (string, error) {
		path, ok := m.Path()
		if !ok {
			return "", errors.New("query path is not known")
		}
		filename := filepath.Base(path)
		filebasename := strings.TrimSuffix(filename, filepath.Ext(filename))
		return toSnake(filebasename), nil
	}
)
