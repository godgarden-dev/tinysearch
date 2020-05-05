package tinysearch

import (
	"fmt"
	"math"
	"sort"
)

// SearchTopK の検索結果を保持する
type TopDocs struct {
	totalHits int         // ヒット件数
	scoreDocs []*ScoreDoc // 検索結果
}

func (t *TopDocs) String() string {
	return fmt.Sprintf("\ntotal hits: %v\nresults: %v\n",
		t.totalHits, t.scoreDocs)
}

// ドキュメントID そのドキュメントのスコアを保持する
type ScoreDoc struct {
	docID DocumentID
	score float64
}

func (d *ScoreDoc) String() string {
	return fmt.Sprintf("docId: %v, Score: %v", d.docID, d.score)
}

type Searcher struct {
	indexReader *IndexReader // インデックス読み込み器
	cursors     []*Cursor    // ポスティングリストのポインタの配列
}

func NewSearcher(path string) *Searcher {
	return &Searcher{indexReader: NewIndexReader(path)}
}

func (s *Searcher) SearchTopK(query []string, k int) *TopDocs {
	// マッチするドキュメントを抽出しスコアを計算する
	results := s.search(query)

	// 結果をスコアの降順でソートする
	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	total := len(results)
	if len(results) > k {
		results = results[:k] // 上位k件のみ取得
	}

	return &TopDocs{
		totalHits: total,
		scoreDocs: results,
	}
}

func (s *Searcher) search(query []string) []*ScoreDoc {

	// カーソルの取得
	// クエリに含まれる用語のポスティングリストが
	// ひとつも存在しない場合、0件え終了する
	if s.openCursors(query) == 0 {
		return []*ScoreDoc{}
	}

	// 一番短いポスティングリストを参照するカーソルを選択
	c := s.cursors[0]
	cursors := s.cursors[1:]

	// 結果を格納する構造体の初期化
	docs := make([]*ScoreDoc, 0)

	// 最も短いポスティングリストをたどり終えるまで繰り返す
	for !c.Empty() {

		var nextDocID DocumentID

		// その他のカーロスをcのdocID以上になるまで進める
		for _, cursor := range cursors {
			if cursor.NextDoc(c.DocId()); cursor.Empty() {
				return docs
			}
			// docId が一致しなければ
			if cursor.DocId() != c.DocId() {
				nextDocId = cursor.DocId()
				break
			}
		}

		if nectDocId > 0 {
			// nextDocId 以上になるまで読みすすめる
			if c.NextDoc(nextDocID); c.Empty() {
				return docs
			}
		} else {
			// 結果を格納
			docs = append(docs, &ScoreDoc{
				docID: c.DocId(),
				score: s.calcScore(),
			})
			c.Next()
		}
	}

	return docs
}

// 検索に使用するポスティングリストのポインタを取得する
// 作成したカーソルの数を返す
func (s *Searcher) openCursors(query []string) int {

	// ポスティングリストを取得
	postings := s.indexReader.postingsList(query)
	if len(postings) == 0 {
		return 0
	}
	// ポスティングリストの短い順にソート
	sort.Slice(postings, func(i, j int) bool {
		return postings[i].Len() < postings[j].Len()
	})
	// 各ポスティングリストに対する cursor の取得
	cursors := make([]*Cursor, len(postings))
	for i, postingList := range postings {
		cursors[i] = postingList.OpenCursor()
	}
	s.cursors = cursors
	return len(cursors)
}

// tf-idf スコアを計算する
func (s *Searcher) calcScore() float64 {
	var score float64
	for i := 0; i < len(s.cursors); i++ {
		termFreq := s.cursors[i].Posting().TermFrequency
		docCount := s.cursors[i].postingsList.Len()
		totalDocCount := s.indexReader.totalDocCount()
		score += calcTF(termFreq) * calIDF(totalDocCount, docCount)
	}
	return score
}

// TF の計算
func calcTF(termCount int) float64 {
	if termCount <= 0 {
		return 0
	}
	return math.Log2(float64(termCount)) + 1
}

// IDF の計算
func calIDF(N, df int) float64 {
	return math.Log2(float64(N) / float64(df))
}
