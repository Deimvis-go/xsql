package sqlfw

import (
	"embed"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/Deimvis/go-ext/go1.25/xembed"
)

//go:embed test_data
var testData embed.FS

func TestReadFromFs(t *testing.T) {
	const expOk bool = true
	const expFail bool = false
	type opts = []FsReadOption
	type expQueries = map[string]string // field name -> field value
	type expMeta = map[string]QueryMeta // field name -> meta
	type QM interface {
		ReadFromFS(fs xembed.Fs, basePath string, opts ...FsReadOption) (FsReadStats, error)
	}
	tcs := []struct {
		title      string
		qm         QM
		basePath   string
		opts       []FsReadOption
		expOk      bool
		expQueries expQueries
		expMeta    expMeta
	}{
		{
			"empty",
			&StructQueryManager[struct{}]{},
			"test_data/read_from_fs/dir1",
			nil,
			expOk,
			expQueries{},
			expMeta{},
		},
		{
			"non-existent-path",
			&StructQueryManager[struct {
				QueryNotFound string `sql:"non_existent.sql"`
			}]{},
			"test_data/read_from_fs/dir1",
			nil,
			expFail,
			expQueries{},
			expMeta{
				"QueryA": queryMeta{name: ""},
			},
		},
		{
			"empty-tag/default-opts",
			&StructQueryManager[struct {
				QueryA string ``
			}]{},
			"test_data/read_from_fs/dir1",
			nil,
			expFail,
			expQueries{},
			expMeta{
				"QueryA": queryMeta{name: ""},
			},
		},
		{
			"path-from-tag/default-opts",
			&StructQueryManager[struct {
				QueryA string `sql:"path=a.sql"`
			}]{},
			"test_data/read_from_fs/dir1",
			nil,
			expOk,
			expQueries{
				"QueryA": `sql query a`,
			},
			expMeta{
				"QueryA": queryMeta{name: ""},
			},
		},
		{
			"abs-path-from-tag/default-opts", // bad pattern actually, but fun that it works :)
			&StructQueryManager[struct {
				QueryA string `sql:"path=/test_data/read_from_fs/dir1/a.sql"`
			}]{},
			"test_data/read_from_fs/dir1",
			nil,
			expOk,
			expQueries{
				"QueryA": `sql query a`,
			},
			expMeta{
				"QueryA": queryMeta{name: ""},
			},
		},
		{
			"no-path-in-tag/default-opts",
			&StructQueryManager[struct {
				QueryA string `sql:""`
			}]{},
			"test_data/read_from_fs/dir1",
			nil,
			expFail,
			expQueries{},
			expMeta{
				"QueryA": queryMeta{name: ""},
			},
		},
		{
			"path-from-field-name/opts",
			&StructQueryManager[struct {
				A string ``
			}]{},
			"test_data/read_from_fs/dir1",
			opts{WithQueryPathResolver(QueryPathFromFieldName)},
			expOk,
			expQueries{
				"A": `sql query a`,
			},
			expMeta{
				"A": queryMeta{name: ""},
			},
		},
		{
			"name-from-tag/default-opts",
			&StructQueryManager[struct {
				QueryA string `sql:"name=a"`
			}]{},
			"test_data/read_from_fs/dir1",
			opts{WithRequireAllQueriesFound(false)},
			expOk,
			expQueries{},
			expMeta{
				"QueryA": queryMeta{name: "a"},
			},
		},
		{
			"name-from-field-name/opts",
			&StructQueryManager[struct {
				QueryA string `sql:""`
			}]{},
			"test_data/read_from_fs/dir1",
			opts{WithRequireAllQueriesFound(false), WithQueryNameResolver(QueryNameFromFieldName)},
			expOk,
			expQueries{},
			expMeta{
				"QueryA": queryMeta{name: "query_a"},
			},
		},
		{
			"name-from-file-basename/opts",
			&StructQueryManager[struct {
				QueryA string `sql:"path=a.sql"`
			}]{},
			"test_data/read_from_fs/dir1",
			opts{WithRequireAllQueriesFound(false), WithQueryNameResolver(QueryNameFromFileBasename)},
			expOk,
			expQueries{},
			expMeta{
				"QueryA": queryMeta{name: "a"},
			},
		},
		{
			"field-Query",
			&StructQueryManager[struct {
				QueryA Query `sql:"path=a.sql;name=a"`
			}]{},
			"test_data/read_from_fs/dir1",
			opts{},
			expOk,
			expQueries{
				"QueryA": `sql query a`,
			},
			expMeta{
				"QueryA": queryMeta{name: "a"},
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.title, func(t *testing.T) {
			_, err := tc.qm.ReadFromFS(testData, tc.basePath, tc.opts...)
			if tc.expOk {
				require.NoError(t, err)
				queries := reflect.ValueOf(tc.qm).MethodByName("Queries").Call(nil)[0].Elem()
				for field, expContent := range tc.expQueries {
					qField := queries.FieldByName(field)
					if qField.Kind() == reflect.Invalid {
						panic(fmt.Errorf("bad test case: no field with name '%s'", field))
					}
					var actContent string
					f := queries.FieldByName(field)
					if f.Type() == queryType {
						actContent = f.Interface().(Query).Raw
					} else {
						actContent = f.String()
					}
					require.Equal(t, expContent, actContent)
				}
				for field, expMeta := range tc.expMeta {
					qField := queries.FieldByName(field)
					if qField.Kind() == reflect.Invalid {
						panic(fmt.Errorf("bad test case: no field with name '%s'", field))
					}
					ret := reflect.ValueOf(tc.qm).MethodByName("MetaOf").Call([]reflect.Value{qField.Addr()})
					actMeta := ret[0].Interface()
					ok := ret[1]
					require.Equal(t, true, ok.Bool())
					require.Equal(t, expMeta, actMeta)
				}
			} else {
				require.Error(t, err)
			}
		})
	}
}
