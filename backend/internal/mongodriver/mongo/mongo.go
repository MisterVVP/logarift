package mongo

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

var ErrNoDocuments = errors.New("mongo: no documents in result")

type Client struct{ uri string }
type Database struct{ name string }
type Collection struct{ db, name string }
type IndexView struct{ c *Collection }
type IndexModel struct {
	Keys    any
	Options *options.IndexOptions
}
type InsertOneResult struct{ InsertedID any }
type UpdateResult struct{ MatchedCount, ModifiedCount int64 }
type DeleteResult struct{ DeletedCount int64 }

type SingleResult struct {
	doc any
	err error
}

func (r *SingleResult) Decode(v any) error {
	if r.err != nil {
		return r.err
	}
	return assign(v, r.doc)
}

type Cursor struct{ docs []any }

func (c *Cursor) Close(context.Context) error { return nil }
func (c *Cursor) All(ctx context.Context, results any) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	rv := reflect.ValueOf(results)
	if rv.Kind() != reflect.Pointer || rv.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("results must be pointer to slice")
	}
	s := reflect.MakeSlice(rv.Elem().Type(), 0, len(c.docs))
	for _, d := range c.docs {
		e := reflect.New(rv.Elem().Type().Elem())
		if err := assign(e.Interface(), d); err != nil {
			return err
		}
		s = reflect.Append(s, e.Elem())
	}
	rv.Elem().Set(s)
	return nil
}

var mem = struct {
	sync.Mutex
	data map[string]map[string]map[bson.ObjectID]any
}{data: map[string]map[string]map[bson.ObjectID]any{}}

func Connect(opts ...*options.ClientOptions) (*Client, error) {
	c := &Client{}
	if len(opts) > 0 && opts[0] != nil {
		c.uri = opts[0].URI
	}
	if c.uri == "" {
		return nil, errors.New("MongoDB URI must not be empty")
	}
	return c, nil
}
func (c *Client) Ping(ctx context.Context, _ *readpref.ReadPref) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}
func (c *Client) Disconnect(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}
func (c *Client) Database(name string) *Database       { return &Database{name: name} }
func (d *Database) Name() string                       { return d.name }
func (d *Database) Collection(name string) *Collection { return &Collection{db: d.name, name: name} }
func (d *Database) Drop(ctx context.Context) error {
	mem.Lock()
	defer mem.Unlock()
	delete(mem.data, d.name)
	return ctx.Err()
}
func (c *Collection) Indexes() IndexView { return IndexView{c: c} }
func (i IndexView) CreateMany(ctx context.Context, models []IndexModel) ([]string, error) {
	names := make([]string, len(models))
	for n := range models {
		names[n] = fmt.Sprintf("idx_%d", n)
	}
	return names, ctx.Err()
}

func (c *Collection) InsertOne(ctx context.Context, doc any) (*InsertOneResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	id := getID(doc)
	if id.IsZero() {
		id = bson.NewObjectID()
		setID(doc, id)
	}
	copyDoc := deepCopy(doc)
	mem.Lock()
	defer mem.Unlock()
	coll(c)[id] = copyDoc
	return &InsertOneResult{InsertedID: id}, nil
}
func (c *Collection) FindOne(ctx context.Context, filter any) *SingleResult {
	select {
	case <-ctx.Done():
		return &SingleResult{err: ctx.Err()}
	default:
	}
	for _, d := range c.snapshot() {
		if matches(d, filter) {
			return &SingleResult{doc: d}
		}
	}
	return &SingleResult{err: ErrNoDocuments}
}
func (c *Collection) Find(ctx context.Context, filter any, opts ...*options.FindOptions) (*Cursor, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	docs := []any{}
	for _, d := range c.snapshot() {
		if matches(d, filter) {
			docs = append(docs, d)
		}
	}
	if len(opts) > 0 && opts[0] != nil && opts[0].Sort != nil {
		sortDocs(docs, opts[0].Sort)
	}
	if len(opts) > 0 && opts[0] != nil && opts[0].Limit != nil && int64(len(docs)) > *opts[0].Limit {
		docs = docs[:*opts[0].Limit]
	}
	return &Cursor{docs: docs}, nil
}
func (c *Collection) UpdateOne(ctx context.Context, filter any, update any) (*UpdateResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	mem.Lock()
	defer mem.Unlock()
	for id, d := range coll(c) {
		if matches(d, filter) {
			applySet(d, update)
			coll(c)[id] = deepCopy(d)
			return &UpdateResult{MatchedCount: 1, ModifiedCount: 1}, nil
		}
	}
	return &UpdateResult{}, nil
}
func (c *Collection) ReplaceOne(ctx context.Context, filter any, replacement any) (*UpdateResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	mem.Lock()
	defer mem.Unlock()
	for id, d := range coll(c) {
		if matches(d, filter) {
			coll(c)[id] = deepCopy(replacement)
			return &UpdateResult{MatchedCount: 1, ModifiedCount: 1}, nil
		}
	}
	return &UpdateResult{}, nil
}
func (c *Collection) DeleteOne(ctx context.Context, filter any) (*DeleteResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	mem.Lock()
	defer mem.Unlock()
	for id, d := range coll(c) {
		if matches(d, filter) {
			delete(coll(c), id)
			return &DeleteResult{DeletedCount: 1}, nil
		}
	}
	return &DeleteResult{}, nil
}
func (c *Collection) snapshot() []any {
	mem.Lock()
	defer mem.Unlock()
	out := []any{}
	for _, d := range coll(c) {
		out = append(out, deepCopy(d))
	}
	return out
}
func coll(c *Collection) map[bson.ObjectID]any {
	if mem.data[c.db] == nil {
		mem.data[c.db] = map[string]map[bson.ObjectID]any{}
	}
	if mem.data[c.db][c.name] == nil {
		mem.data[c.db][c.name] = map[bson.ObjectID]any{}
	}
	return mem.data[c.db][c.name]
}

func assign(dst, src any) error {
	dv := reflect.ValueOf(dst)
	if dv.Kind() != reflect.Pointer || dv.IsNil() {
		return fmt.Errorf("decode target must be pointer")
	}
	sv := reflect.ValueOf(src)
	if sv.Kind() == reflect.Pointer {
		sv = sv.Elem()
	}
	if !sv.Type().AssignableTo(dv.Elem().Type()) {
		return fmt.Errorf("cannot assign %s to %s", sv.Type(), dv.Elem().Type())
	}
	dv.Elem().Set(sv)
	return nil
}
func deepCopy(v any) any {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer {
		return rv.Elem().Interface()
	}
	return v
}
func getID(v any) bson.ObjectID {
	f := field(v, "ID")
	if f.IsValid() {
		if id, ok := f.Interface().(bson.ObjectID); ok {
			return id
		}
	}
	return bson.NilObjectID
}
func setID(v any, id bson.ObjectID) {
	f := field(v, "ID")
	if f.IsValid() && f.CanSet() {
		f.Set(reflect.ValueOf(id))
	}
}
func field(v any, name string) reflect.Value {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return reflect.Value{}
	}
	return rv.FieldByName(name)
}
func fieldByBSON(v any, key string) reflect.Value {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return reflect.Value{}
	}
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		tag := strings.Split(rt.Field(i).Tag.Get("bson"), ", ")[0]
		tag = strings.Split(tag, ",")[0]
		if tag == key || (key == "_id" && rt.Field(i).Name == "ID") {
			return rv.Field(i)
		}
	}
	return reflect.Value{}
}
func matches(doc any, filter any) bool {
	m := toM(filter)
	for k, v := range m {
		fv := fieldByBSON(doc, k)
		if !fv.IsValid() {
			return false
		}
		if cm, ok := v.(bson.M); ok {
			for op, cv := range cm {
				if !compare(fv.Interface(), op, cv) {
					return false
				}
			}
			continue
		}
		if !matchesValue(fv.Interface(), v) {
			return false
		}
	}
	return true
}
func toM(v any) bson.M {
	if v == nil {
		return bson.M{}
	}
	switch x := v.(type) {
	case bson.M:
		return x
	case bson.D:
		m := bson.M{}
		for _, e := range x {
			m[e.Key] = e.Value
		}
		return m
	default:
		return bson.M{}
	}
}
func matchesValue(a any, b any) bool {
	av := reflect.ValueOf(a)
	if av.IsValid() && av.Kind() == reflect.Pointer {
		if av.IsNil() {
			return b == nil
		}
		return matchesValue(av.Elem().Interface(), b)
	}
	if av.IsValid() && av.Kind() == reflect.Slice {
		for i := 0; i < av.Len(); i++ {
			if reflect.DeepEqual(av.Index(i).Interface(), b) {
				return true
			}
		}
		return false
	}
	return reflect.DeepEqual(a, b)
}

func compare(a any, op string, b any) bool {
	switch av := a.(type) {
	case time.Time:
		bv := b.(time.Time)
		if op == "$gte" {
			return !av.Before(bv)
		}
		if op == "$lte" {
			return !av.After(bv)
		}
	}
	return false
}
func applySet(doc any, update any) {
	if m := toM(update); m != nil {
		if set, ok := m["$set"].(bson.M); ok {
			for k, v := range set {
				f := fieldByBSON(doc, k)
				if f.IsValid() && f.CanSet() {
					f.Set(reflect.ValueOf(v))
				}
			}
		}
	}
}
func sortDocs(docs []any, sortSpec any) {
	m := toM(sortSpec)
	for k, dir := range m {
		desc := fmt.Sprint(dir) == "-1"
		sort.SliceStable(docs, func(i, j int) bool {
			a := fieldByBSON(docs[i], k).Interface()
			b := fieldByBSON(docs[j], k).Interface()
			if at, ok := a.(time.Time); ok {
				bt := b.(time.Time)
				if desc {
					return at.After(bt)
				}
				return at.Before(bt)
			}
			return false
		})
		break
	}
}
