package parser

type Node interface {
	iNode()
}

type SayHelloCommand struct {
	Node
}

type ShowCommand struct {
	Node
	Type string
}
