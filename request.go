package server

const (
	FILE_CREATE uint32 = iota
	FILE_UNLINK
	DIR_LIST
)

type Request struct {
	Op   uint32
	Name string
}
