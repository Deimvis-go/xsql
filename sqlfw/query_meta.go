package sqlfw

type QueryMeta interface {
	// Name will return empty string "" if query name is unknown.
	Name() string
}

// TODO: support ad-hoc resolving
// (e.g. if QueryNameResolver calls Path() and path is not known yet,
// but QueryPathResolver exists, then it is invoked, but algorithm
// should prevent call cycles)
type QueryFieldMeta interface {
	Name() (string, bool)
	Path() (string, bool)
	FieldName() string
	FieldTag() parsedTag
}

func newQueryFieldMeta(fieldName string, fieldTag parsedTag) *queryFieldMeta {
	return &queryFieldMeta{
		fieldName: fieldName,
		fieldTag:  fieldTag,
	}
}

type queryMeta struct {
	name string
}

func (qm queryMeta) Name() string {
	return qm.name
}

type queryFieldMeta struct {
	name      optional[string]
	path      optional[string]
	fieldName string
	fieldTag  parsedTag
}

func (qfm *queryFieldMeta) Name() (string, bool) {
	if qfm.name.HasValue() {
		return qfm.name.Value(), true
	}
	return "", false
}
func (qfm *queryFieldMeta) Path() (string, bool) {
	if qfm.path.HasValue() {
		return qfm.path.Value(), true
	}
	return "", false
}
func (qfm *queryFieldMeta) FieldName() string {
	return qfm.fieldName
}

func (qfm *queryFieldMeta) FieldTag() parsedTag {
	return qfm.fieldTag
}
