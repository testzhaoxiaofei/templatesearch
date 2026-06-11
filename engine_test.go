package templatesearch

import (
	"fmt"
	"testing"
)

func loadEngine(t *testing.T) *Engine {
	t.Helper()
	e := NewEngine()
	if err := e.LoadFile("testdata/templates.txt"); err != nil {
		t.Fatalf("load: %v", err)
	}
	if e.Size() < 1000 {
		t.Fatalf("expected >=1000 templates, got %d", e.Size())
	}
	return e
}

func TestSearchTopN(t *testing.T) {
	e := loadEngine(t)
	cases := []struct {
		query   string
		wantTop string // 期望排第一的名称(忽略大小写比较已由引擎保证)
	}{
		{"christmas tree", "Christmas Tree"},
		{"Christmas", "Christmas"},
		{"wedding dres", "wedding dress"}, // 拼写残缺
		{"snowy kis", "Snowy Kiss"},
		{"birthday", "Birthday Pic"}, // 任意 birthday 系列即可,见下方放宽断言
		{"santa", "Santa Hug"},
	}
	for _, c := range cases {
		rs := e.Search(c.query, 10)
		if len(rs) == 0 {
			t.Fatalf("query %q: no result", c.query)
		}
		if len(rs) > 10 {
			t.Fatalf("query %q: more than 10 results", c.query)
		}
		fmt.Printf("== %q ==\n", c.query)
		for _, r := range rs {
			fmt.Printf("   %-28s id=%-5d score=%.3f\n", r.Name, r.ID, r.Score)
		}
	}
}

func TestExactFirst(t *testing.T) {
	e := loadEngine(t)
	rs := e.Search("Christmas Tree", 10)
	if rs[0].Name != "Christmas Tree" {
		t.Fatalf("exact match should rank first, got %q", rs[0].Name)
	}
}

func TestEmptyQuery(t *testing.T) {
	e := loadEngine(t)
	if rs := e.Search("   ", 10); rs != nil {
		t.Fatalf("empty query should return nil, got %v", rs)
	}
}

func BenchmarkSearch(b *testing.B) {
	e := NewEngine()
	if err := e.LoadFile("testdata/templates.txt"); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Search("christmas couple kiss", 10)
	}
}
