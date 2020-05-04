package tinysearch

import (
	"database/sql"
	"io"
	"os"
	"path/filepath"
)

type Engine struct {
	tokenizer     *Tokenizer     // トークンを分割する
	indexer       *Indexer       // インデックスを作成する
	documentStore *DocumentStore // ドキュメントを管理する
	indexDir      string         // インデックスファイルを保存するディレクトリ
}

// 検索エンジンを作成する処理
func NewSearchEngine(db *sql.DB) *Engine {

	tokenizer := NewTokenizer()
	indexer := NewIndexer(tokenizer)
	documentStore := NewDocumentStore(db)

	path, ok := os.LookupEnv("INDEX_DIR_PATH")
	if !ok {
		current, _ := os.Getwd()
		path = filepath.Join(current, "_index_data")
	}
	return &Engine{
		tokenizer:     tokenizer,
		indexer:       indexer,
		documentStore: documentStore,
		indexDir:      path,
	}
}

// インデックスにドキュメントを追加する
func (e *Engine) AddDocument(title string, reader io.Reader) error {
	id, err := e.documentStore.save(title) // タイトルを保存しドキュメントIDを発行する

	if err != nil {
		return err
	}
	e.indexer.update(id, reader) // インデックスを更新する
	return nil
}

// インデックスをファイルに書き出す
func (e *Engine) Flush() error {
	writer := NewIndexWriter(e.indexDir)
	return writer.Flush(e.indexer.index)
}
