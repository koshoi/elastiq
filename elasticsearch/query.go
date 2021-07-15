package elasticsearch

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
	Debug     bool
	Recursive bool
	Raw       bool
	AsCurl    bool
}
