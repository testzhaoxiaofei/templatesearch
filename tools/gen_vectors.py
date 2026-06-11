#!/usr/bin/env python3
"""
离线生成模版向量(JSONL),供 Go 服务的 VectorIndex 加载。

用法:
  pip install FlagEmbedding            # 自建 bge-m3 方案
  python tools/gen_vectors.py search.txt vectors.jsonl

bge-m3 支持 100+ 语言,模版名是英文、查询是任意语言也能在同一向量空间对齐。
如改用云端 embedding API,把 embed_batch 替换为对应 SDK 调用即可。
模版数据更新后重跑本脚本并调 POST /api/templates/reload(需同时重载向量,
或直接滚动重启服务)。
"""
import json
import sys


def load_templates(path):
    items = []
    with open(path, encoding="utf-8") as f:
        for line in f:
            line = line.strip()
            if not line or ":" not in line:
                continue
            name, _, id_ = line.rpartition(":")
            try:
                items.append((int(id_), name.strip()))
            except ValueError:
                continue
    return items


def embed_batch(texts):
    from FlagEmbedding import BGEM3FlagModel
    model = BGEM3FlagModel("BAAI/bge-m3", use_fp16=True)
    out = model.encode(texts, batch_size=64)["dense_vecs"]
    return out.tolist()


def main():
    src = sys.argv[1] if len(sys.argv) > 1 else "search.txt"
    dst = sys.argv[2] if len(sys.argv) > 2 else "vectors.jsonl"
    items = load_templates(src)
    print(f"loaded {len(items)} templates from {src}")
    vecs = embed_batch([name for _, name in items])
    with open(dst, "w", encoding="utf-8") as f:
        for (id_, name), v in zip(items, vecs):
            f.write(json.dumps(
                {"id": id_, "name": name, "vector": [round(x, 6) for x in v]},
                ensure_ascii=False) + "\n")
    print(f"wrote {len(items)} vectors to {dst}")


if __name__ == "__main__":
    main()
