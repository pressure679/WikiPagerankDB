package djikstra

import (
	"container/heap"
	"math"
)

// edge struct holds the bare data needed to define a graph.
type edge struct {
	vert1, vert2 string
	dist         int
}

func NewEdge(newvert1, newvert2 string, newdist int) edge {
	return edge{
		vert1: newvert1,
		vert2: newvert2,
		dist: newdist,
	}
}

/*
// NewEdge(vert1, vert2, dist, graph)
// allNodes, startNode, endNode := LinkGraph(graph, directed, start, end)
// paths := Dijkstra(allNodes, startNode, endNode)

func main() {
	// example data and parameters
	graph := []edge{
		{"a", "b", 7},
		{"a", "c", 9},
		{"a", "f", 14},
		{"b", "c", 10},
		{"b", "d", 15},
		{"c", "d", 11},
		{"c", "f", 2},
		{"d", "e", 6},
		{"e", "f", 9},
	}
	directed := true
	start := "a"
	end := "e"
	findAll := false

	// construct linked representation of example data
	allNodes, startNode, endNode := linkGraph(graph, directed, start, end)
	if directed {
		fmt.Print("Directed")
	} else {
		fmt.Print("Undirected")
	}
	fmt.Printf(" graph with %d nodes, %d edges\n", len(allNodes), len(graph))
	if startNode == nil {
		fmt.Printf("start node %q not found in graph\n", start)
		return
	}
	if findAll {
		endNode = nil
	} else if endNode == nil {
		fmt.Printf("end node %q not found in graph\n", end)
		return
	}

	// run Dijkstra's shortest path algorithm
	paths := dijkstra(allNodes, startNode, endNode)
	fmt.Println("Shortest path(s):")
	for _, p := range paths {
		fmt.Println(p.path, "length", p.length)
	}
}
*/

type node struct {
	vert string     // vertex name
	tent int        // tentative distance
	prev *node      // previous node in shortest path back to start
	done bool       // true when tent and prev represent shortest path
	nbs  []neighbor // edges from this vertex
	rx   int        // heap.Remove index
}

type neighbor struct {
	nd   *node // node corresponding to vertex
	dist int   // distance to this node (from whatever node references this)
}

func LinkGraph(graph []edge, directed bool,
	start, end string) (allNodes []*node, startNode, endNode *node) {

	all := make(map[string]*node)
	// one pass over graph to collect nodes and link neighbors
	for _, e := range graph {
		n1 := all[e.vert1]
		n2 := all[e.vert2]
		// add previously unseen nodes
		if n1 == nil {
			n1 = &node{vert: e.vert1}
			all[e.vert1] = n1
		}
		if n2 == nil {
			n2 = &node{vert: e.vert2}
			all[e.vert2] = n2
		}
		// link neighbors
		n1.nbs = append(n1.nbs, neighbor{n2, e.dist})
		if !directed {
			n2.nbs = append(n2.nbs, neighbor{n1, e.dist})
		}
	}
	allNodes = make([]*node, len(all))
	var n int
	for _, nd := range all {
		allNodes[n] = nd
		n++
	}
	return allNodes, all[start], all[end]
}

// return type
type path struct {
	path   []string
	length int
}

func Dijkstra(allNodes []*node, startNode, endNode *node) (pl []path) {
	// WP steps 1 and 2.
	for _, nd := range allNodes {
		nd.tent = math.MaxInt32
		nd.done = false
		nd.prev = nil
		nd.rx = -1
	}
	current := startNode
	current.tent = 0
	var unvis ndList

	for {
		// WP step 3: update tentative distances to neighbors
		for _, nb := range current.nbs {
			if nd := nb.nd; !nd.done {
				if d := current.tent + nb.dist; d < nd.tent {
					nd.tent = d
					nd.prev = current
					if nd.rx < 0 {
						heap.Push(&unvis, nd)
					} else {
						heap.Fix(&unvis, nd.rx)
					}
				}
			}
		}
		// WP step 4: mark current node visited, record path and distance
		current.done = true
		if endNode == nil || current == endNode {
			// record path and distance for return value
			distance := current.tent
			// recover path by tracing prev links,
			var p []string
			for ; current != nil; current = current.prev {
				p = append(p, current.vert)
			}
			// then reverse list
			for i := (len(p) + 1) / 2; i > 0; i-- {
				p[i-1], p[len(p)-i] = p[len(p)-i], p[i-1]
			}
			pl = append(pl, path{p, distance}) // pl is return value
			// WP step 5 (case of end node reached)
			if endNode != nil {
				return
			}
		}
		if len(unvis) == 0 {
			break // WP step 5 (case of no more reachable nodes)
		}
		// WP step 6: new current is node with smallest tentative distance
		current = heap.Pop(&unvis).(*node)
	}
	return
}

// ndList implements container/heap
type ndList []*node

func (n ndList) Len() int           { return len(n) }
func (n ndList) Less(i, j int) bool { return n[i].tent < n[j].tent }
func (n ndList) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
	n[i].rx = i
	n[j].rx = j
}
func (n *ndList) Push(x interface{}) {
	nd := x.(*node)
	nd.rx = len(*n)
	*n = append(*n, nd)
}
func (n *ndList) Pop() interface{} {
	s := *n
	last := len(s) - 1
	r := s[last]
	*n = s[:last]
	r.rx = -1
	return r
}
