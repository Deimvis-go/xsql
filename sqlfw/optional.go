package sqlfw

type optional[U any] struct {
	v   U
	set bool
}

func (o optional[U]) HasValue() bool { return o.set }
func (o optional[U]) Value() U       { return o.v }
func (o *optional[U]) SetValue(v U)  { o.v = v; o.set = true }

func optionalValueOr[U any](o optional[U], fb U) U {
	if o.set {
		return o.v
	}
	return fb
}
