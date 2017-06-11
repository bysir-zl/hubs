package message

type Interface interface {
	Marshal() (bs []byte, err error)
	UnMarshal(bs []byte) (err error)
}
