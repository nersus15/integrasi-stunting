package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"sync"

	"github.com/webcore-go/webcore/port"
)

type InsertOneCall struct {
	Table string
	Data  any
}

type FindCall struct {
	Table  string
	Filter []port.DbExpression
}

// Expr mengembalikan ekspresi filter pertama, bagian yang benar-benar disambung
// ke string SQL oleh lib-sql. Nilai argumen tidak pernah masuk ke sini.
func (c FindCall) Expr() string {
	if len(c.Filter) == 0 {
		return ""
	}
	return c.Filter[0].Expr
}

type DeleteCall struct {
	Table  any
	Filter []port.DbExpression
}

type UpdateCall struct {
	Table  string
	Filter []port.DbExpression
	Data   any
}

// FakeDatabase meniru port.IDatabase tanpa database sungguhan.
//
// InsertOne mengembalikan argumen data apa adanya, persis seperti lib-sql yang
// dipakai di produksi. Ini penting: kalau repository mengasumsikan tipe lain,
// test harus ikut gagal seperti produksi.
type FakeDatabase struct {
	mu sync.Mutex

	InsertOneErr       error
	InsertOneFunc      func(ctx context.Context, table string, data any) (any, error)
	FindFunc           func(ctx context.Context, results any, table string, filter []port.DbExpression) error
	FindDeletedFunc    func(ctx context.Context, results any, table string, filter []port.DbExpression) error
	FindOneFunc        func(ctx context.Context, result any, table string, filter []port.DbExpression) error
	FindOneDeletedFunc func(ctx context.Context, result any, table string, filter []port.DbExpression) error
	DeleteErr          error
	UpdateOneAffected  int64
	UpdateOneErr       error

	InsertOneCalls []InsertOneCall
	FindCalls      []FindCall
	DeleteCalls    []DeleteCall
	UpdateOneCalls []UpdateCall
}

var _ port.IDatabase = (*FakeDatabase)(nil)

func NewFakeDatabase() *FakeDatabase { return &FakeDatabase{} }

func NewFailingFakeDatabase(err error) *FakeDatabase {
	return &FakeDatabase{InsertOneErr: err}
}

func (f *FakeDatabase) InsertOne(ctx context.Context, table string, data any) (any, error) {
	f.mu.Lock()
	f.InsertOneCalls = append(f.InsertOneCalls, InsertOneCall{Table: table, Data: data})
	fn, err := f.InsertOneFunc, f.InsertOneErr
	f.mu.Unlock()

	if fn != nil {
		return fn(ctx, table, data)
	}
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (f *FakeDatabase) Find(ctx context.Context, results any, table string, column []string, filter []port.DbExpression, sort map[string]int, limit int64, skip int64) error {
	f.mu.Lock()
	f.FindCalls = append(f.FindCalls, FindCall{Table: table, Filter: filter})
	fn := f.FindFunc
	f.mu.Unlock()

	if fn != nil {
		return fn(ctx, results, table, filter)
	}
	return nil
}
func (f *FakeDatabase) FindDeleted(ctx context.Context, results any, table string, column []string, filter []port.DbExpression, sort map[string]int, limit int64, skip int64) error {
	f.mu.Lock()
	f.FindCalls = append(f.FindCalls, FindCall{Table: table, Filter: filter})
	fn := f.FindDeletedFunc
	f.mu.Unlock()

	if fn != nil {
		return fn(ctx, results, table, filter)
	}
	return nil
}
func (f *FakeDatabase) DeleteOne(ctx context.Context, table any, filter []port.DbExpression) (int64, error) {
	f.mu.Lock()
	f.DeleteCalls = append(f.DeleteCalls, DeleteCall{Table: table, Filter: filter})
	err := f.DeleteErr
	f.mu.Unlock()

	if err != nil {
		return 0, err
	}
	return 1, nil
}

func (f *FakeDatabase) InsertOneCallCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.InsertOneCalls)
}

func (f *FakeDatabase) LastInsertOne() (InsertOneCall, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if len(f.InsertOneCalls) == 0 {
		return InsertOneCall{}, false
	}
	return f.InsertOneCalls[len(f.InsertOneCalls)-1], true
}

// DeletedTables mengembalikan tabel yang pernah dihapus, berurutan.
func (f *FakeDatabase) DeletedTables() []string {
	f.mu.Lock()
	defer f.mu.Unlock()

	table := make([]string, 0, len(f.DeleteCalls))
	for _, c := range f.DeleteCalls {
		tname := ""
		switch reflect.TypeOf(c.Table).Kind() {
		case reflect.Ptr, reflect.Struct:
			tname = GetTableName(c.Table)
		default:
			tname = reflect.ValueOf(c.Table).String()

		}

		table = append(table, tname)
	}

	return table
}

func GetTableName(v any) string {
	rv := reflect.ValueOf(v)

	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return ""
		}
		rv = rv.Elem()
	}

	method := rv.MethodByName("TableName")
	if !method.IsValid() {
		return ""
	}

	results := method.Call(nil)
	if len(results) == 0 {
		return ""
	}

	name, ok := results[0].Interface().(string)
	if !ok {
		return ""
	}

	return name
}

// --- sisa kontrak port.IDatabase, tidak dipakai alur create ---

func (f *FakeDatabase) Install(args ...any) error      { return nil }
func (f *FakeDatabase) Uninstall() error               { return nil }
func (f *FakeDatabase) Connect() error                 { return nil }
func (f *FakeDatabase) Disconnect() error              { return nil }
func (f *FakeDatabase) Ping(ctx context.Context) error { return nil }
func (f *FakeDatabase) GetConnection() any             { return nil }
func (f *FakeDatabase) GetDriver() string              { return "fake" }
func (f *FakeDatabase) GetName() string                { return "fake-db" }

func (f *FakeDatabase) Count(ctx context.Context, table string, filter []port.DbExpression) (int64, error) {
	return 0, fmt.Errorf("Count tidak diimplementasikan pada FakeDatabase")
}

func (f *FakeDatabase) FindOne(ctx context.Context, result any, table string, column []string, filter []port.DbExpression, sort map[string]int) error {
	f.mu.Lock()
	f.FindCalls = append(f.FindCalls, FindCall{Table: table, Filter: filter})
	fn := f.FindOneFunc
	f.mu.Unlock()

	if fn != nil {
		return fn(ctx, result, table, filter)
	}
	// Tanpa FindOneFunc, perlakukan seperti tidak ada baris - sama seperti bun.
	return sql.ErrNoRows
}

func (f *FakeDatabase) FindOneDeleted(ctx context.Context, result any, table string, column []string, filter []port.DbExpression, sort map[string]int) error {
	f.mu.Lock()
	f.FindCalls = append(f.FindCalls, FindCall{Table: table, Filter: filter})
	fn := f.FindOneDeletedFunc
	f.mu.Unlock()

	if fn != nil {
		return fn(ctx, result, table, filter)
	}
	// Tanpa FindOneFunc, perlakukan seperti tidak ada baris - sama seperti bun.
	return sql.ErrNoRows
}

func (f *FakeDatabase) Update(ctx context.Context, table string, filter []port.DbExpression, data any) (int64, error) {
	return 0, fmt.Errorf("Update tidak diimplementasikan pada FakeDatabase")
}

// UpdateOne mencatat panggilan dan mengembalikan UpdateOneAffected baris
func (f *FakeDatabase) UpdateOne(ctx context.Context, table string, filter []port.DbExpression, data any) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.UpdateOneCalls = append(f.UpdateOneCalls, UpdateCall{Table: table, Filter: filter, Data: data})
	if f.UpdateOneErr != nil {
		return 0, f.UpdateOneErr
	}
	return f.UpdateOneAffected, nil
}

func (f *FakeDatabase) Delete(ctx context.Context, table any, filter []port.DbExpression) (int64, error) {
	return 0, fmt.Errorf("Delete tidak diimplementasikan pada FakeDatabase")
}

func (f *FakeDatabase) StartMigration(ctx context.Context, service string, command string, dir string, args []string) error {
	return nil
}
