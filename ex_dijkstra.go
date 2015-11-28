//    This code is an example of djikstra's algorithm by the help of the library I modified from http://rosettacode.org/wiki/Dijkstra's_algorithm#Go to fit as a Go package
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
package main
import (
	"fmt"
	"github.com/pressure679/dijkstra"
)
func main() {
	// Making edges
	edge1 := dijkstra.NewEdge("a", "b", 1)
	edge2 := dijkstra.NewEdge("b", "c", 1)
	edges := make([]dijkstra.Edge, 2)
	edges[0] = edge1
	edges[1] = edge2

	// Linking edges to nodes, argument false/true is whether graph is directed or not
	allNodes, startNode, endNode := dijkstra.LinkGraph(edges, false, "a", "c")
	fmt.Println("Graph linked")
	
	// Printing nodes, neighbors and edges
	fmt.Println(allNodes)
	fmt.Printf("\nStart node\n")
	fmt.Printf("vert:%s, tent:%d, done:%t, heap.remove_index:%d\n", startNode.Vert, startNode.Tent, startNode.Done, startNode.Rx)
	if startNode.Prev != nil {
		fmt.Printf("neighbor:%s\n", startNode.Prev.Vert)
	}
	if startNode.Nbs != nil {
		for _, nb := range(startNode.Nbs) {
			fmt.Printf("startNode.Nbs:%s\n", nb.Nd.Vert)
		}
	}
	fmt.Printf("\nEnd node\n")
	fmt.Printf("vert:%s, tent:%d, done:%t, heap.remove_index:%d\n", endNode.Vert, endNode.Tent, endNode.Done, endNode.Rx)
	if endNode.Prev != nil {
		fmt.Printf("neighbor:%s\n", endNode.Prev.Vert)
	}
	if endNode.Nbs != nil {
		for _, nb := range(endNode.Nbs) {
			fmt.Printf("endNode.Nbs:%s\n", nb.Nd.Vert)
		}
	}
	fmt.Printf("\nEdges\n")
	fmt.Printf("edge1.Vert1:%s, edge1.Vert2:%s, edge1.Dist:%d\n", edge1.Vert1, edge1.Vert2, edge1.Dist)
	fmt.Printf("edge2.Vert1:%s, edge2.Vert2:%s, edge2.Dist:%d\n", edge2.Vert1, edge2.Vert2, edge2.Dist)
	fmt.Printf("\nGraph connected\n")

	// Calculating the path with dijkstra algorithm and printing the path
	path := dijkstra.Dijkstra(allNodes, startNode, endNode)
	for _, node := range(path) {
		for num, nodename := range(node.Path) {
			if num < len(node.Path) - 1 {
				fmt.Printf("[num:%d, name:%s] -> ", num, nodename)
			} else {
				fmt.Printf("[num:%d, name:%s]\n", num, nodename)
			}
		}
		fmt.Printf("length:%d\n", node.Length)
	}
	fmt.Println()
}
