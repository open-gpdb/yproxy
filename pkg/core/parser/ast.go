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

type KKBCommand struct {
	Node
}

type CopyCommand struct {
	Node
	Path    string
	Options []Node
}

type Option struct {
	Node
	Name string
	Arg  Node
}

type AExprSConst struct {
	Node
	Value string
}

type AExprIConst struct {
	Node
	Value int
}

type AExprBConst struct {
	Node
	Value bool
}
