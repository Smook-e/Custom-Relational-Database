package linkedlist

import (
	"fmt"
)

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
func AddNodetoHead(head **Node, value any) {
	newNode := NewNode(value)
	if *head == nil {
		*head = newNode
		return
	}

	newNode.Next = *head
	(*head).prev = newNode
	*head = newNode
}

func RemoveTailFromHead(head **Node) {
	if *head == nil {
		return
	}

	if (*head).Next == nil {
		*head = nil
		return
	}

	current := *head
	for current.Next != nil {
		current = current.Next
	}
	current.prev.Next = nil
}

func RemoveNode(head **Node, nodeToRemove *Node) {
	if *head == nil || nodeToRemove == nil {
		return
	}

	if *head == nodeToRemove {
		*head = nodeToRemove.Next
		if *head != nil {
			(*head).prev = nil
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
func FindNode(head *Node, value any) *Node {
	current := head
	for current != nil {
		if current.Value == value {
			return current
		}
		current = current.Next
	}
	return nil
}

func PrintList(head *Node) {
	current := head
	for current != nil {
		fmt.Println(current.Value)
		current = current.Next
	}
}
func PrintListReverse(tail *Node) {
	current := tail
	for current != nil {
		fmt.Println(current.Value)
		current = current.prev
	}
}