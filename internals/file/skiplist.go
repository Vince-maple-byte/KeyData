package file

import "math/rand/v2"

var MAX_LEVEL int = 32; 

type Node struct {
	key string
	levels []*Node
}

type Skiplist struct {
	head *Node
	
}

func randLevel() int {
	var level int = 1;

	for(rand.IntN(2) == 1 && level < MAX_LEVEL){
		level++;
	}

	return level;
}
