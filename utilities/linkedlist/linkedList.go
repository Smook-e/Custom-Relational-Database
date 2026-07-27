package linkedlist

import (
	"fmt"
)

type Node struct {
	Value any
	PageID uint32
	Next  *Node
	prev  *Node
}
type LinkedList struct {
	Head *Node
	Tail *Node
}

func NewNode(value any) *Node {
	return &Node{Value: value}
}

func (L *LinkedList) AddNodetoTail(value any) {
	newNode := NewNode(value)
	if L.Tail == nil {
		L.Head = newNode
		L.Tail = newNode
		return
	}

	
	L.Tail.Next = newNode
	newNode.prev = L.Tail
	L.Tail = newNode
}
func (L *LinkedList) AddNodetoHead(value any) {
	newNode := NewNode(value)
	if L.Head == nil {
		L.Head = newNode
		L.Tail = newNode
		return
	}

	newNode.Next = L.Head
	L.Head.prev = newNode
	L.Head = newNode
}
func (L *LinkedList) RemoveTail() *Node {
	if L.Tail == nil {
		
		return nil
	}

	if L.Tail.prev == nil {
		n := L.Tail
		L.Tail = nil
		return n
	}
	n := L.Tail
	L.Tail = L.Tail.prev
	L.Tail.Next = nil
	return n
}

func (L *LinkedList) RemoveHead() *Node{
	if L.Head == nil {
		return nil
	}

	if L.Head.Next == nil {
		n := L.Head
		L.Head = nil
		return n
	}
	n := L.Head
	L.Head = L.Head.Next
	L.Head.prev = nil
	return n
}

func (L *LinkedList) RemoveNode(nodeToRemove *Node) {
	if L.Head == nil || nodeToRemove == nil {
		return
	}

	if L.Head == nodeToRemove {
		L.Head = nodeToRemove.Next
		if L.Head != nil {
			L.Head.prev = nil
		}
		return
	}

	if nodeToRemove.prev != nil {
		nodeToRemove.prev.Next = nodeToRemove.Next
	}
	if nodeToRemove.Next != nil {
		nodeToRemove.Next.prev = nodeToRemove.prev
	}
}
func (L *LinkedList) MoveNodeToHead(nodeToMove *Node) {
	if L.Head == nil || nodeToMove == nil || L.Head == nodeToMove {
		return
	}

	if L.Tail == nodeToMove {
		L.Tail = nodeToMove.prev
	}

	// Remove the node from its current position
	if nodeToMove.prev != nil {
		nodeToMove.prev.Next = nodeToMove.Next
	}
	if nodeToMove.Next != nil {
		nodeToMove.Next.prev = nodeToMove.prev
	}
	// Move the node to the head
	nodeToMove.Next = L.Head
	nodeToMove.prev = nil
	L.Head.prev = nodeToMove
	L.Head = nodeToMove

}
func (L *LinkedList) MoveNodeToTail(nodeToMove *Node) {
	if L.Tail == nil || nodeToMove == nil || L.Tail == nodeToMove {
		return
	}

	if L.Head == nodeToMove {
		L.Head = nodeToMove.Next
	}

	// Remove the node from its current position
	if nodeToMove.prev != nil {
		nodeToMove.prev.Next = nodeToMove.Next
	}
	if nodeToMove.Next != nil {
		nodeToMove.Next.prev = nodeToMove.prev
	}

	// Move the node to the tail
	nodeToMove.prev = L.Tail
	nodeToMove.Next = nil
	L.Tail.Next = nodeToMove
	L.Tail = nodeToMove
}
func (L *LinkedList) FindNodeByValue(value any) *Node {
	current := L.Head
	for current != nil {
		if current.Value == value {
			return current
		}
		current = current.Next
	}
	return nil
}
func (L *LinkedList) FindNode(value any) *Node {
	current := L.Head
	for current != nil {
		if current.Value == value {
			return current
		}
		current = current.Next
	}
	return nil
}

func (L *LinkedList) PrintList() {
	current := L.Head
	for current != nil {
		fmt.Println(current.Value)
		current = current.Next
	}
}
func (L *LinkedList) PrintListReverse() {
	current := L.Tail
	for current != nil {
		fmt.Println(current.Value)
		current = current.prev
	}
}

func (L *LinkedList) IsEmpty() bool {
	return L.Head == nil
}