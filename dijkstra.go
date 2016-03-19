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
	"os"
	"bufio"
	"fmt"
	"strings"
	"strconv"
	"bytes"
)
// edge struct holds the bare data needed to define a graph.
type Node struct {
	Vert string     // vertex name
	Nbs  []Neighbor // edges from this vertex
	tent int        // tentative distance
	prev *Node      // previous node in shortest path back to start
	done bool       // true when tent and prev represent shortest path
	rx   int        // heap.Remove index
}
type Neighbor struct {
	Nd   *Node // node corresponding to vertex
	Dist int   // distance to this node (from whatever node references this)
}
func (graph *Node) AppendNeighbor(neighbor Neighbor) {
	graph.Nbs = append(graph.Nbs, neighbor)
}
type Path struct {
	Path   []string
	Length int
}
func LinkGraph(graph map[string]*Neighbor /*, start, end string */) (allNodes []*Node /*, startNode, endNode *Node */) {
	// one pass over graph to collect nodes and link neighbors
	/* for node, _ := range(graph) {
		// add previously unseen nodes
		if graph[node] == nil {
			graph[node] = &Node{Vert: neighbors.Vert}
		}

		// link neighbors
		for _, neighbor := range(graph[node].Nbs) {
			graph[node].Nbs = append(graph[node].Nbs, Neighbor{graph[neighbor.Vert], 1})
			// if !directed {
			graph[neighbor.Vert].Nbs = append(graph[neighbor].Nbs, Neighbor{graph[node.Vert], 1})
		}
	} */
	allNodes = make([]*Node, len(graph))
	var n int = 1
	
	for _, nd := range(graph) {
		allNodes[n] = nd.Nd
		n++
	}
	return allNodes /*, all[start], all[end] */
}
func Dijkstra(allNodes []*Node, startNode, endNode *Node) (pl []Path) {
	// WP steps 1 and 2.
	for _, node := range allNodes {
		node.tent = math.MaxInt32
		node.done = false
		node.prev = nil
		node.rx = -1
	}
	current := startNode
	current.tent = 0
	var unvis ndList

	for {
		// WP step 3: update tentative distances to neighbors
		for _, nb := range(current.Nbs) {
			if nd := nb.Nd; !nd.done {
				if d := current.tent + nb.Dist; d < nd.tent {
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
		// WP step 4: mark startNode node visited, record path and distance
		startNode.done = true
		if endNode == nil || startNode == endNode {
			// record path and distance for return value
			distance := startNode.tent
			// recover path by tracing prev links,
			var p []string
			for ; startNode != nil; startNode = startNode.prev {
				p = append(p, startNode.Vert)
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
		// WP step 6: new startNode is node with smallest tentative distance
		startNode = heap.Pop(&unvis).(*Node)
	}
	return
}
// ndList implements container/heap
type ndList []*Node
func (n ndList) Len() int           { return len(n) }
func (n ndList) Less(i, j int) bool { return n[i].tent < n[j].tent }
func (n ndList) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
	n[i].rx = i
	n[j].rx = j
}
func (n *ndList) Push(x interface{}) {
	nd := x.(*Node)
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
func CreateDB(Graph []*Node, FileName string) error {
	file, err := os.Open("dijkstra.dat")
	if err != nil {
		return err
	}
	defer file.Close()
	fwriter := bufio.NewWriter(file)
	for _, node := range(Graph) {
		fmt.Fprintln(fwriter, node.Vert)
		for _, neighbor := range(node.Nbs) {
			fmt.Fprintln(fwriter, " " + strconv.Itoa(neighbor.Dist) + "-" + neighbor.Nd.Vert)
		}
	}
	return nil
}
func ReadDB(FileName string) (err error, Graph []*Node) {
	file, err := os.Open(FileName)
	if err != nil {
		return err, nil
	}
	fReader := bufio.NewReader(file)
	var node *Node
	var neighbor *Neighbor
	strSplit := make([]string, 2)
	var sliceToString bytes.Buffer
	for {
		line, _, err := fReader.ReadLine()
		if err != nil {
			return err, nil
		}
		sliceToString.Write(line)
		strLine := sliceToString.String()
		if !strings.EqualFold(strLine[0:1], " ") {
			node.Vert = strLine
		} else {
			for strings.EqualFold(strLine[:1], " ") {
				strSplit = strings.Split(strLine[1:], "-")
				distInt64, err  := strconv.ParseInt(strSplit[0], 10, 0)
				if err != nil {
					return err, nil
				}
				neighbor.Dist = int(distInt64)
				neighbor.Nd.Vert = strSplit[1]
				fReader.ReadLine()
				node.AppendNeighbor(*neighbor)
			}
		}
		Graph = append(Graph, node)
	}
	return
}
