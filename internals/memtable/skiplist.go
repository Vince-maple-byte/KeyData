package memtable

import (
	"errors"
	"math/rand/v2"
)

const MAX_LEVEL = 32; 

type Node struct {
	key string
	value []byte
	levels []*Node
}

type Skiplist struct {
	head *Node
}

func CreateSkiplist() *Skiplist {
	head := &Node{
		key: "",
		value: nil,
		levels: make([]*Node, MAX_LEVEL),
	};
	return &Skiplist{head: head};
}

func (list *Skiplist) Search(key string) ([]byte, error) {
  	curr := list.head;
	var next *Node;

	for i := len(list.head.levels)-1; i >= 0; i-- {
		
		for next = curr.levels[i]; next != nil && key > next.key; next = curr.levels[i] {
			curr = next;
		}
		if next != nil && next.key == key {
			return next.value, nil;
		}
	} 

	return nil, errors.New("Element doesn't exist");
}

func (list *Skiplist) Insert(key string, value []byte) {
	var update [MAX_LEVEL]*Node;
	curr := list.head;

	for i:= MAX_LEVEL-1; i >= 0; i-- {
		for curr.levels[i] != nil && key > curr.levels[i].key {
			curr = curr.levels[i];
		}

		update[i] = curr;
	}
	
	//This code block checks if the key is already inside of the skiplist
	next := curr.levels[0];
	if next != nil && next.key == key {
		next.value = value;
		return;
	}

	newNode := &Node{
		key: key, 
		value: value,
		levels: make([]*Node, randLevel()),
	}

	for i := 0; i < len(newNode.levels); i++ {
		newNode.levels[i] = update[i].levels[i];
		update[i].levels[i] = newNode;
	}
}

func (list *Skiplist) Delete(key string) ([]byte,error) {
	deleteNode, err := list.searchNode(key);

	if err != nil {
		return nil, err;
	}

	update := make([]*Node, len(deleteNode.levels));
	curr := list.head;
	val := deleteNode.value;

	for i:= len(update)-1; i>=0; i-- {
		for curr.levels[i] != nil && key > curr.levels[i].key {
    		curr = curr.levels[i]
		}	

		update[i] = curr;
	}

	for i:= 0; i < len(update); i++ {
		update[i].levels[i] = deleteNode.levels[i];
	}

	return val, nil;
}

func (list Skiplist) searchNode(key string) (*Node, error) {
	curr := list.head;
	var next *Node;

	for i := len(list.head.levels)-1; i >= 0; i-- {
		
		for next = curr.levels[i]; next != nil && key > next.key; next = curr.levels[i] {
			curr = next;
		}
		if next != nil && next.key == key {
			return next, nil;
		}
	} 

	return nil, errors.New("Element doesn't exist");
}

func (list *Skiplist) EmptyList() {
	list = CreateSkiplist();
}

func randLevel() int {
	var level int = 1;

	for(rand.IntN(2) == 1 && level < MAX_LEVEL){
		level++;
	}

	return level;
}
