package elasticsearch

type FilterOperation string

const (
	EQ  FilterOperation = "eq"
	NEQ FilterOperation = "neq"
	GT  FilterOperation = "gt"
	GTE FilterOperation = "gte"
	LT  FilterOperation = "lt"
	LTE FilterOperation = "lte"
	IN  FilterOperation = "in"
	LK  FilterOperation = "lk"
	BT  FilterOperation = "bt"
)

type Filter struct {
	Key       string
	Value     []string
	Operation FilterOperation
}

type Order struct {
	By        string
	Ascending bool
}

type Query struct {
	Filters []*Filter
	Order   *Order
	Limit   int
	Index   string
	Output  string
}

type Options struct {
	Debug  bool
	Raw    bool
	AsCurl bool
}
