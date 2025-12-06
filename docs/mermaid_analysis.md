# Mermaid 関連の把握メモ

## 1. Mermaid 文字列生成に関与するファイル
- `cmd/marid/main.go`：CLI エントリーポイント。接続・スキーマ抽出後に `diagram.Generate` を呼び、生成した Mermaid 文字列を標準出力へ流す。
- `internal/diagram/generate.go`：Mermaid ER 図の文字列を構築するレンダラー本体。スキーマ情報から `erDiagram` セクション、テーブル定義、リレーション行を生成する。
- `internal/diagram/generate_test.go`：`Generate` の出力を直接比較するユニットテスト。
- `internal/schema/extract.go`：DB からスキーマ（テーブル、カラム、キー定義、コメント）を取得するフォーマット非依存のレイヤー。Mermaid 生成に必要なデータのソース。
- `pkg/utils/schema_utils.go`：識別子の整形・型の表示用フォーマッタ。Mermaid 等のダイアグラム出力向けに利用できるが、現状では他コードから未使用。

## 2. Mermaid 固有ロジックとフォーマット非依存ロジックの分類
### Mermaid 固有
- `internal/diagram/generate.go` の `Generate`：Mermaid の構文 (`erDiagram`、エンティティブロック、`||--o{` など) を直接組み立てる。
- 同ファイル内の `Relationship` 構造体と整列ロジック：Mermaid のエッジ出力順を制御するためのもの。
- `cmd/marid/main.go` の標準出力処理：Mermaid 文字列をそのまま出力する動線。

### フォーマット非依存
- `internal/schema/extract.go` と関連関数：データベースからスキーマ情報を集約するだけで、出力フォーマットには依存しない。
- `pkg/utils/schema_utils.go`：識別子・型の整形ユーティリティ。ダイアグラム出力向けだがフォーマットを限定せず再利用可能。
- `internal/diagram/generate.go` 内の列挙・ソート前処理（テーブル走査や距離計算などのロジック部分）：データ構造の準備が中心で、Mermaid 固有の文字列化とは分離可能。

## 3. Mermaid 文字列を直接検証しているテスト
- `internal/diagram/generate_test.go`：`expected` 文字列として完全な Mermaid ER 図を組み立て、`Generate` の返却値と厳密比較している。抽象化後は期待値の組み立て方法または比較方法の見直しが必要な候補。

## 4. このメモをリポジトリに残す理由
- Mermaid 生成部の抽象化や出力フォーマット拡張を検討する際、既存の依存関係とテストの影響範囲を素早く把握できるリファレンスになる。
- CLI・レンダラー・スキーマ抽出の役割分担を明文化することで、別タスクで担当者が変わっても同じ前提に立てる。
- 直接文字列比較しているテストの一覧を共有しておくことで、後続のリファクタリング時に「どこを先に直すか」を判断しやすくなる。
