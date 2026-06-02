package options

type ClientOptions struct{ URI string }

func Client() *ClientOptions                                { return &ClientOptions{} }
func (o *ClientOptions) ApplyURI(uri string) *ClientOptions { o.URI = uri; return o }

type FindOptions struct {
	Limit *int64
	Sort  any
}

func Find() *FindOptions                             { return &FindOptions{} }
func (o *FindOptions) SetLimit(v int64) *FindOptions { o.Limit = &v; return o }
func (o *FindOptions) SetSort(v any) *FindOptions    { o.Sort = v; return o }

type IndexOptions struct {
	Unique                  *bool
	Name                    *string
	PartialFilterExpression any
}

func Index() *IndexOptions                             { return &IndexOptions{} }
func (o *IndexOptions) SetUnique(v bool) *IndexOptions { o.Unique = &v; return o }
func (o *IndexOptions) SetName(v string) *IndexOptions { o.Name = &v; return o }
func (o *IndexOptions) SetPartialFilterExpression(v any) *IndexOptions {
	o.PartialFilterExpression = v
	return o
}
