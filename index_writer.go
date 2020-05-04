package tinysearch

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

type IndexWriter struct {
	indexDir string
}

func NewIndexWriter(path string) *IndexWriter  {
	return &IndexWriter{path}
}

// インデックスの永続化処理
func (w *IndexWriter) Flush(index *Index) error  {
	for term, postingsList := range index.Dictionary {
		if err := w.postingsList(term, postingsList); err != nil {
			fmt.Printf("failed to save %s postings list: %v", term, err)
		}
	}
	return w.docCount(index.TotalDocsCount)
}

func (w *IndexWriter) postingsList(term string, list PostingsList) error {
	bytes, err := json.Marshal(list)
	if err != nil {
		return err
	}

	filename := filepath.Join(w.indexDir, term)
	file, err := os.Create(filename)
	if err != nil {
		return nil
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	_, err = writer.Write(bytes)
	if err != nil {
		return err
	}
	return writer.Flush()
}

func (w *IndexWriter) docCount(count int) error {
	filename := filepath.Join(w.indexDir, "_0.dc")
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write([]byte(strconv.Itoa(count)))
	return err
}

