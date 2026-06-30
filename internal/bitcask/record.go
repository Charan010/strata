package bitcask

type Record struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	Timestamp int64  `json:"timestamp"`
	Deleted   bool   `json:"deleted"`
}
