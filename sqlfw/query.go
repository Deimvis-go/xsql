package sqlfw

type Query struct {
	Raw  string
	Meta QueryMeta
}

// TODO: impement injection-safe formatting,
// may be useful for debugging purposes
// func (q Query) Format(args... any)
