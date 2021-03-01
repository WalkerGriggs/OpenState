package openstate

type MessageType uint8

const (
	NameAddRequestType MessageType = 0
)

type NameAddRequest struct {
	Name string
}
