package snf

type Operation uint32

const (
	FileCreate Operation = iota
	FileUnlink
	DirList
)

type Request struct {
	Op   Operation
	Name string
}
