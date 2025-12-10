# Formatter testing template

新しいフォーマットを追加する際に利用できるテスト用テンプレートです。`pkg/formatter/formattertest` パッケージに用意したモックやサンプルデータを活用し、既存の契約テストにスムーズに組み込めるようにします。

## 1. サンプルデータの利用

Mermaid など既存フォーマッタが利用する最小限のスキーマは `formattertest.SampleRenderData()` で取得できます。新フォーマッタのレンダリングテストでもこのデータを使えば、主キー・ユニークキー・外部キーを含む典型ケースをカバーできます。

```go
import (
    "testing"

    "github.com/motchang/marid/pkg/formatter/formattertest"
)

func TestRenderMyFormat(t *testing.T) {
    data := formattertest.SampleRenderData()
    // テスト対象のフォーマッタでレンダリング
}
```

## 2. モックフォーマッタの活用

フォーマッタを DI するハンドラやレジストリの振る舞いを検証したい場合は `formattertest.MockFormatter` を利用できます。呼び出しデータの記録や任意の戻り値設定が可能です。

```go
mock := &formattertest.MockFormatter{
    NameValue:      "my-format",
    MediaTypeValue: "text/x-my-format",
    RenderFunc: func(data formatter.RenderData) (string, error) {
        // 呼び出し内容を検証しつつ任意の文字列を返却
        return "dummy-output", nil
    },
}
```

`RenderCalls` に渡されたデータが順次蓄積されるため、テスト完了後に呼び出し回数やパラメータを確認できます。

## 3. 契約テストへの追加

`pkg/formatter/formatter_contract_test.go` にフォーマッタ固有の期待値を 1 ケース追加するだけで、共通の契約 (Name / MediaType / Render 出力) を検証できます。新しいフォーマッタを実装したら、以下のようにエントリを足してください。

```go
{
    name:            "my-format implements contract",
    formatter:       myformat.New(),
    wantName:        "my-format",
    wantMediaType:   "text/x-my-format",
    wantRenderMatch: "...期待する文字列...",
},
```

この契約テストは、フォーマッタがインターフェース要件を満たし、既存フォーマットと同等の振る舞いをするかを素早く検証するためのベースラインとなります。
