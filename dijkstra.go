//    This is a modified version of the one from http://rosettacode.org/wiki/Dijkstra's_algorithm#Go , and it is modified to fit as a Go package.
//    Copyright (C) 2015  Vittus Peter Ove Maqe Mikiassen
//
//    This program is free software: you can redistribute it and/or modify
//    it under the terms of the GNU General Public License as published by
//    the Free Software Foundation, either version 3 of the License, or
//    (at your option) any later version.
//
//    This program is distributed in the hope that it will be useful,
//    but WITHOUT ANY WARRANTY; without even the implied warranty of
//    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//    GNU General Public License for more details.
//
//    You should have received a copy of the GNU General Public License
//    along with this program.  If not, see <http://www.gnu.org/licenses/>.
package dijkstra
import (
	"container/heap"
	"math"
)
// edge struct holds the bare data needed to define a graph.
type Edge struct {
	Vert1, Vert2 string
	Dist         int
}
func NewEdge(newvert1, newvert2 string, newdist int) Edge {
	return Edge{
		Vert1: newvert1,
		Vert2: newvert2,
		Dist: newdist,
	}
}
func LinkGraph(graph []Edge, directed bool, start, end string) (allNodes []*Node, startNode, endNode *Node) {
	all := make(map[string]*Node)
	// one pass over graph to collect nodes and link neighbors
	for _, e := range graph {
		n1 := all[e.Vert1]
		n2 := all[e.Vert2]
		// add previously unseen nodes
		if n1 == nil {
			n1 = &Node{Vert: e.Vert1}
			all[e.Vert1] = n1
		}
		if n2 == nil {
			n2 = &Node{Vert: e.Vert2}
			all[e.Vert2] = n2
		}
		// link neighbors
		n1.Nbs = append(n1.Nbs, Neighbor{n2, e.Dist})
		if !directed {
			n2.Nbs = append(n2.Nbs, Neighbor{n1, e.Dist})
		}
	}
	allNodes = make([]*Node, len(all))
	var n int
	for _, nd := range all {
		allNodes[n] = nd
		n++
	}
	return allNodes, all[start], all[end]
}
func Dijkstra(allNodes []*Node, startNode, endNode *Node) (pl []Path) {
	// WP steps 1 and 2.
	for _, nd := range allNodes {
		nd.Tent = math.MaxInt32
		nd.Done = false
		nd.Prev = nil
		nd.Rx = -1
	}
	current := startNode
	current.Tent = 0
	var unvis ndList

	for {
		// WP step 3: update tentative distances to neighbors
		for _, nb := range current.Nbs {
			if nd := nb.Nd; !nd.Done {
				if d := current.Tent + nb.Dist; d < nd.Tent {
					nd.Tent = d
					nd.Prev = current
					if nd.Rx < 0 {
						heap.Push(&unvis, nd)
					} else {
						heap.Fix(&unvis, nd.Rx)
					}
				}
			}
		}
		// WP step 4: mark current node visited, record path and distance
		current.Done = true
		if endNode == nil || current == endNode {
			// record path and distance for return value
			distance := current.Tent
			// recover path by tracing prev links,
			var p []string
			for ; current != nil; current = current.Prev {
				p = append(p, current.Vert)
			}
			// then reverse list
			for i := (len(p) + 1) / 2; i > 0; i-- {
				p[i-1], p[len(p)-i] = p[len(p)-i], p[i-1]
			}
			pl = append(pl, Path{p, distance}) // pl is return value
			// WP step 5 (case of end node reached)
			if endNode != nil {
				return
			}
		}
		if len(unvis) == 0 {
			break // WP step 5 (case of no more reachable nodes)
		}
		// WP step 6: new current is node with smallest tentative distance
		current = heap.Pop(&unvis).(*Node)
	}
	return
}
type Node struct {
	Vert string     // vertex name
	Tent int        // tentative distance
	Prev *Node      // previous node in shortest path back to start
	Done bool       // true when tent and prev represent shortest path
	Nbs  []Neighbor // edges from this vertex
	Rx   int        // heap.Remove index
}
type Neighbor struct {
	Nd   *Node // node corresponding to vertex
	Dist int   // distance to this node (from whatever node references this)
}
// return type
type Path struct {
	Path   []string
	Length int
}

// ndList implements container/heap
type ndList []*Node
func (n ndList) Len() int           { return len(n) }
func (n ndList) Less(i, j int) bool { return n[i].Tent < n[j].Tent }
func (n ndList) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
	n[i].Rx = i
	n[j].Rx = j
}
func (n *ndList) Push(x interface{}) {
	nd := x.(*Node)
	nd.Rx = len(*n)
	*n = append(*n, nd)
}
func (n *ndList) Pop() interface{} {
	s := *n
	last := len(s) - 1
	r := s[last]
	*n = s[:last]
	r.Rx = -1
	return r
}

/*
// NewEdge(vert1, vert2, dist, graph)
// allNodes, startNode, endNode := LinkGraph(graph, directed, start, end)
// paths := Dijkstra(allNodes, startNode, endNode)

func main() {
	// example data and parameters
	graph := []Edge{
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
