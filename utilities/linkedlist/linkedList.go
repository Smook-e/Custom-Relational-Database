package linkedlist



type Node struct {
	Value any
	Next  *Node
	prev  *Node
}

func NewNode(value any) *Node {
	return &Node{Value: value}
}

func AddNodetoTail(head **Node, value any) {
	newNode := NewNode(value)
	if *head == nil {
		*head = newNode
		return
	}

	current := *head
	for current.Next != nil {
		current = current.Next
	}
	current.Next = newNode
	newNode.prev = current
}