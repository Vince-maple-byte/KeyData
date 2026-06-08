package memtable

import (
	"errors"
	"math/rand/v2"

	"github.com/Vince-maple-byte/KeyData/internals/record"
)

const MAX_LEVEL = 32

type Node struct {
	key    string
	value  []byte
	levels []*Node
}

type Skiplist struct {
	head *Node
	size int
}

func CreateSkiplist() *Skiplist {
	head := &Node{
		key:    "",
		value:  nil,
		levels: make([]*Node, MAX_LEVEL),
	}
	return &Skiplist{
		head: head,
		size: 0,
	}
}

func (list *Skiplist) Search(key string) ([]byte, error) {
	curr := list.head
	var next *Node

	for i := len(list.head.levels) - 1; i >= 0; i-- {

		for next = curr.levels[i]; next != nil && key > next.key; next = curr.levels[i] {
			curr = next
		}
		if next != nil && next.key == key {
			return next.value, nil
		}
	}

	return nil, errors.New("Element doesn't exist")
}

// The value is the entire record including all of the information on the timestamp, etc.
func (list *Skiplist) Insert(key string, value []byte) {
	var update [MAX_LEVEL]*Node
	curr := list.head

	for i := MAX_LEVEL - 1; i >= 0; i-- {
		for curr.levels[i] != nil && key > curr.levels[i].key {
			curr = curr.levels[i]
		}

		update[i] = curr
	}

	//This code block checks if the key is already inside of the skiplist
	// If the key is already inside then we can check for whichever key is the newest by time and then have that value in the skiplist
	next := curr.levels[0]
	if next != nil && next.key == key {
		newTime := record.GetContents(value).Timestamp.UTC()
		prevTime := record.GetContents(next.value).Timestamp.UTC()

		if newTime.After(prevTime) || newTime.Equal(prevTime) {
			next.value = value
		}

		return
	}

	newNode := &Node{
		key:    key,
		value:  value,
		levels: make([]*Node, randLevel()),
	}

	for i := 0; i < len(newNode.levels); i++ {
		newNode.levels[i] = update[i].levels[i]
		update[i].levels[i] = newNode
	}

	list.size++
}

func (list *Skiplist) Delete(key string) ([]byte, error) {
	deleteNode, err := list.searchNode(key)

	if err != nil {
		return nil, err
	}

	update := make([]*Node, len(deleteNode.levels))
	curr := list.head
	val := deleteNode.value

	for i := len(update) - 1; i >= 0; i-- {
		for curr.levels[i] != nil && key > curr.levels[i].key {
			curr = curr.levels[i]
		}

		update[i] = curr
	}

	for i := range len(update) {
		update[i].levels[i] = deleteNode.levels[i]
	}

	list.size--
	return val, nil
}

func (list Skiplist) searchNode(key string) (*Node, error) {
	curr := list.head
	var next *Node

	for i := len(list.head.levels) - 1; i >= 0; i-- {

		for next = curr.levels[i]; next != nil && key > next.key; next = curr.levels[i] {
			curr = next
		}
		if next != nil && next.key == key {
			return next, nil
		}
	}

	return nil, errors.New("Element doesn't exist")
}

func (list *Skiplist) EmptyList() {
	// head := &Node{
	// 	key: "",
	// 	value: nil,
	// 	levels: make([]*Node, MAX_LEVEL),
	// };
	// return &Skiplist{
	// 	head: head,
	// 	size: 0,
	// };
	list.head.key = ""
	list.head.value = nil
	list.head.levels = make([]*Node, MAX_LEVEL)

	list.size = 0
}

func (list *Skiplist) EntireList() [][]byte {
	curr := list.head.levels[0]
	res := make([][]byte, 0, list.size)

	for range list.size {
		res = append(res, curr.value)
		curr = curr.levels[0]
	}

	return res
}

func randLevel() int {
	var level int = 1

	for rand.IntN(2) == 1 && level < MAX_LEVEL {
		level++
	}

	return level
}
