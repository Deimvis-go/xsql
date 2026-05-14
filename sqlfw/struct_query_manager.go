package sqlfw

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/Deimvis/go-ext/go1.25/xcheck/xshould"
	"github.com/Deimvis/go-ext/go1.25/xembed"
	"github.com/Deimvis/go-ext/go1.25/xmaps"
	"github.com/Deimvis/go-ext/go1.25/xoptional"
)

func NewStructQueryManager[T any]() *StructQueryManager[T] {
	return &StructQueryManager[T]{ptr2meta: make(map[any]QueryMeta)}
}

type StructQueryManager[T any] struct {
	queries  T
	ptr2meta map[any]QueryMeta
}

var _ QueryManager[any] = (*StructQueryManager[any])(nil)

func (qm *StructQueryManager[T]) QueriesSnapshot() T {
	return qm.queries
}

func (qm *StructQueryManager[T]) Queries() *T {
	return &qm.queries
}

func (qm *StructQueryManager[T]) MetaOf(queryPtr any) (QueryMeta, bool) {
	var rawQueryPtr unsafe.Pointer
	switch qp := queryPtr.(type) {
	case *string:
		rawQueryPtr = unsafe.Pointer(qp)
	case *[]byte:
		rawQueryPtr = unsafe.Pointer(qp)
	case *Query:
		rawQueryPtr = unsafe.Pointer(qp)
	default:
		return nil, false
	}
	v, ok := qm.ptr2meta[rawQueryPtr]
	return v, ok
}

type FsReadStats struct {
	ReadQueryCount  int64
	TotalQueryCount int64
}

type FsReadOption func(*fsReadCfg)

// TODO: support subdirectories and recursive walk
func (qm *StructQueryManager[T]) ReadFromFS(fs xembed.Fs, basePath string, opts ...FsReadOption) (FsReadStats, error) {
	cfg := defaultFsReadCfg
	for _, opt := range opts {
		opt(&cfg)
	}
	if qm.ptr2meta == nil {
		qm.ptr2meta = make(map[any]QueryMeta)
	}

	rfs := xembed.NewRelativeFs(fs)
	rfs.Cd("/")
	rfs.Cd(basePath)
	var stats FsReadStats

	v := reflect.ValueOf(&qm.queries) // use pointer in order to obtain field pointers later
	v = v.Elem()
	vt := v.Type()
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		ft := vt.Field(i)
		if ft.Anonymous {
			// ignore embedded field (field with no name)
			continue
		}
		if !ft.IsExported() {
			// ignore private fields
			continue
		}
		if !f.CanSet() {
			// same as ignoring private fields, but also use it for safety
			continue
		}
		tag := ParseTag(vt.Field(i).Tag.Get("sql"))
		if len(tag.kv) == 0 || xmaps.HasKey(tag.kv, "-") {
			// skip some extra fields, not relevant to query manager
			continue
		}

		kind := ft.Type.Kind()
		isStr := kind == reflect.String
		isBytes := (kind == reflect.Slice && ft.Type.Elem().Kind() == reflect.Uint8)
		isQuery := (kind == reflect.Struct && ft.Type == queryType)
		if isStr || isBytes || isQuery {
			stats.TotalQueryCount++
			fmeta := newQueryFieldMeta(ft.Name, tag)

			path, err := cfg.queryPathResolver(fmeta)
			var content []byte
			if err != nil {
				if cfg.requireAllQueriesFound {
					return stats, fmt.Errorf("resolve query path: %w", err)
				}
			} else {
				fmeta.path.SetValue(path)
				content, err = rfs.ReadFile(path)
				if err != nil {
					return stats, err
				}
				stats.ReadQueryCount++
			}

			name, err := cfg.queryNameResolver(fmeta)
			if err == nil {
				fmeta.name.SetValue(name)
			}

			meta := queryMeta{name: xoptional.ValueOr(fmeta.name, "")}
			if isStr {
				f.SetString(string(content))
			} else if isBytes {
				f.SetBytes(content)
			} else {
				q := Query{
					Raw:  string(content),
					Meta: meta,
				}
				f.Set(reflect.ValueOf(q))
			}
			qm.ptr2meta[f.Addr().UnsafePointer()] = meta
		} else if kind == reflect.Struct && xmaps.HasKey(tag.kv, "cd") {
			rfs.Cd(tag.kv["cd"])
			defer rfs.Cd("..")
			// TODO: support recursive (add interface with readFromFs method (lowercase), check impl and call)
			panic("recursive query manager not implemented yet")
		}

	}

	if cfg.requireAllQueriesFound {
		err := xshould.Eq(stats.ReadQueryCount, stats.TotalQueryCount, "some queries not found")
		if err != nil {
			return stats, err
		}
	}

	return stats, nil
}

func WithQueryPathResolver(r QueryPathResolver) FsReadOption {
	return func(c *fsReadCfg) {
		c.queryPathResolver = r
	}
}

func WithQueryNameResolver(r QueryNameResolver) FsReadOption {
	return func(c *fsReadCfg) {
		c.queryNameResolver = r
	}
}

func WithRequireAllQueriesFound(v bool) FsReadOption {
	return func(c *fsReadCfg) {
		c.requireAllQueriesFound = v
	}
}

type fsReadCfg struct {
	queryPathResolver      QueryPathResolver
	queryNameResolver      QueryNameResolver
	requireAllQueriesFound bool
}

var (
	defaultFsReadCfg = fsReadCfg{
		queryPathResolver:      QueryPathFromFieldTag,
		queryNameResolver:      QueryNameFromFieldTag,
		requireAllQueriesFound: true,
	}

	queryType = reflect.TypeOf(Query{})
)
