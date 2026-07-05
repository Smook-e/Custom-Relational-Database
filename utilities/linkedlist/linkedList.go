package linkedlist

import (
	"fmt"
)

type Node struct {
	Value any
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
func (L *LinkedList) RemoveTail() {
	if L.Tail == nil {
		return
	}

	if L.Tail.prev == nil {
		L.Tail = nil
		return
	}

	L.Tail = L.Tail.prev
	L.Tail.Next = nil
}

func (L *LinkedList) RemoveHead() {
	if L.Head == nil {
		return
	}

	if L.Head.Next == nil {
		L.Head = nil
		return
	}

	L.Head = L.Head.Next
	L.Head.prev = nil
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
