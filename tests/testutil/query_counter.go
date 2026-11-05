package testutil

import (
	"context"
	"fmt"
	"regexp"
	"sync"

	"github.com/uptrace/bun"
	"github.com/webcore-go/webcore/port"
)

// QueryCounter mencatat statement yang dijalankan bun selama Ukur berlangsung.
type QueryCounter struct {
	mu    sync.Mutex
	aktif bool
	query []string
}

func (p *QueryCounter) BeforeQuery(ctx context.Context, _ *bun.QueryEvent) context.Context {
	return ctx
}

func (p *QueryCounter) AfterQuery(_ context.Context, e *bun.QueryEvent) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.aktif {
		p.query = append(p.query, e.Query)
	}
}

// Measure menjalankan fn dan mengembalikan statement yang terjadi selama fn berjalan.
// Pemanggil harus memastikan tidak ada request lain yang berjalan bersamaan.
func (p *QueryCounter) Measure(fn func()) Statement {
	p.mu.Lock()
	p.aktif, p.query = true, nil
	p.mu.Unlock()

	fn()

	p.mu.Lock()
	defer p.mu.Unlock()
	p.aktif = false
	return Statement(p.query)
}

// InstallQueryCounter menempelkan hook ke koneksi bun milik aplikasi; cukup sekali per paket.
func (a *App) InstallQueryCounter() (*QueryCounter, error) {
	lib, ok := a.app.Context.GetDefaultSingletonInstance("database")
	if !ok {
		return nil, fmt.Errorf("database tidak termuat")
	}
	db, ok := lib.(port.IDatabase).GetConnection().(*bun.DB)
	if !ok {
		return nil, fmt.Errorf("koneksi database bukan *bun.DB")
	}
	p := &QueryCounter{}
	db.AddQueryHook(p)
	return p, nil
}

type Statement []string

// Match mengembalikan statement yang cocok dengan pola regexp (tanpa membedakan huruf besar/kecil).
func (s Statement) Match(pattern string) Statement {
	re := regexp.MustCompile("(?is)" + pattern)
	var result Statement
	for _, q := range s {
		if re.MatchString(q) {
			result = append(result, q)
		}
	}
	return result
}
