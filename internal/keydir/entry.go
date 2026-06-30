package keydir

type Entry struct {
	Offset    int64
	Size      uint32
	Timestamp int64
	Deleted   bool
}
