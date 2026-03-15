package file

import "math/rand/v2"

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

func (list *Skiplist) Search(key string) []byte {
	curr := list.head;
	var next *Node;

	for i := len(list.head.levels)-1; i >= 0; i-- {
		
		for next = curr.levels[i]; next != nil && key > next.key; next = curr.levels[i] {
			curr = next;
		}
		if next != nil && next.key == key {
			return next.value;
		}
	} 

	return nil;
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

func (list *Skiplist) Delete(key string) []byte {

	return nil;
}

func randLevel() int {
	var level int = 1;

	for(rand.IntN(2) == 1 && level < MAX_LEVEL){
		level++;
	}

	return level;
}
