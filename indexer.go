package tinysearch

import (
	"bufio"
	"io"
)

type Indexer struct {
	index     *Index
	tokenizer *Tokenizer
}

func NewIndexer(tokenizer *Tokenizer) *Indexer {
	return &Indexer{
		index:     NewIndex(),
		tokenizer: tokenizer,
	}
}

// ドキュメントをインデックスに追加する処理
func (idxr *Indexer) update(docID DocumentID, reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	scanner.Split(idxr.tokenizer.SplitFunc) // ① 分割方法の指定
	var position int

	for scanner.Scan() {
		term := scanner.Text() // ② 用語ごとに読み込み

		// ポスティングリストの更新
		if postingsList, ok := idxr.index.Dictionary[term]; !ok {
			// ③ term をキーとするポスティングリストが存在しない場合

			idxr.index.Dictionary[term] =
				NewPostingsList(NewPosting(docID, position))
		} else {
			// ④ ポスティングリストがすでに存在する場合は追加
			postingsList.Add(NewPosting(docID, position))
		}
		position++
	}

	idxr.index.TotalDocsCount++
}
