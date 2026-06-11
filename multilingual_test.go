package templatesearch

import (
	"context"
	"testing"
)

func TestNormalizeFold(t *testing.T) {
	cases := map[string]string{
		"Café Noël":     "cafe noel",
		"NIÑO pequeño":  "nino pequeno",
		"schöne Größe":  "schone grosse",
		"ＣＨＲＩＳＴＭＡＳ":     "christmas", // 全角
		"São Paulo":     "sao paulo",
		"Crème brûlée!": "creme brulee",
	}
	for in, want := range cases {
		if got := normalize(in); got != want {
			t.Errorf("normalize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNoSpaceScripts(t *testing.T) {
	// 韩文/泰文应走字符 n-gram 路径,可被部分匹配召回
	e := NewEngine()
	e.Rebuild([]Template{
		{ID: 1, Name: "크리스마스 트리"},        // 韩语:圣诞树
		{ID: 2, Name: "สุขสันต์วันเกิด"}, // 泰语:生日快乐
		{ID: 3, Name: "Christmas Tree"},
	})
	if rs := e.Search("크리스마스", 5); len(rs) == 0 || rs[0].ID != 1 {
		t.Fatalf("korean query failed: %v", rs)
	}
	if rs := e.Search("วันเกิด", 5); len(rs) == 0 || rs[0].ID != 2 {
		t.Fatalf("thai query failed: %v", rs)
	}
}

type fakeEmbedder struct{ vec []float32 }

func (f *fakeEmbedder) Embed(_ context.Context, _ string) ([]float32, error) {
	return f.vec, nil
}

func TestVectorIndexAndHybrid(t *testing.T) {
	v := NewVectorIndex()
	v.dim = 3
	v.ids = []int{207, 122, 570}
	v.name = []string{"Christmas Tree", "wedding dress", "Birthday Party"}
	v.vecs = [][]float32{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}}

	// 查询向量靠近 Christmas Tree
	rs := v.Search([]float32{0.9, 0.1, 0}, 2)
	if len(rs) != 2 || rs[0].ID != 207 {
		t.Fatalf("vector search wrong: %v", rs)
	}

	// Hybrid:词面查不到中文"圣诞树",但向量路召回 Christmas Tree
	lex := NewEngine()
	lex.Rebuild([]Template{
		{ID: 207, Name: "Christmas Tree"},
		{ID: 122, Name: "wedding dress"},
		{ID: 570, Name: "Birthday Party"},
	})
	h := &Hybrid{Lexical: lex, Vectors: v, Embedder: &fakeEmbedder{vec: []float32{1, 0, 0}}}
	rs = h.Search(context.Background(), "圣诞树", 10)
	if len(rs) == 0 || rs[0].ID != 207 {
		t.Fatalf("hybrid cross-lingual failed: %v", rs)
	}
}

func TestHybridDegrade(t *testing.T) {
	// 向量索引为空时应降级为纯词面,不报错
	lex := NewEngine()
	lex.Rebuild([]Template{{ID: 1, Name: "Snowy Kiss"}})
	h := &Hybrid{Lexical: lex}
	rs := h.Search(context.Background(), "snowy", 10)
	if len(rs) != 1 || rs[0].ID != 1 {
		t.Fatalf("degrade failed: %v", rs)
	}
}
