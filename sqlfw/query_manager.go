package sqlfw

type QueryManager[T any] interface {
	IConstQueryManager[T]
	Queries() *T
	// MetaOf accepts pointer to query,
	// which can be obtained using Queries() method.
	// If input value does not belong to query pointers,
	// false will be returned.
	MetaOf(any) (QueryMeta, bool)
}

type IConstQueryManager[T any] interface {
	QueriesSnapshot() T
}
