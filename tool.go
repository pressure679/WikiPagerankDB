package main
import (
	"fmt"
	"os"
	"bufio"
	"strings"
	"strconv"
	"runtime"
	"time"

  //  "golang.org/x/net/html"
  //  "github.com/akhenakh/gozim"
  //  "github.com/jaytaylor/html2text"

	"github.com/pressure679/go-wikiparse"

	"github.com/alixaxel/pagerank"
)
const wikipedia_dir string = "/home/pressure679/Documents/Wiki/Pedia/"
const wikipedia_index_dir string = "/home/pressure679/wd/Wikipedia_index"
var Titles []string
// var Offsets []uint
type Article struct {
	FileOffsets []uint32
	Links []*string
	Pagerank float64
}
var Articles []Article
// var articles_by_title map[string][]*uint
// var articles_by_offset map[*Offset]*string
var articles_by_title map[string]*Article
var articles_by_offset map[uint32]*string // for pagerank
var wiki_file *os.File
var index_file *os.File
// var graph *pagerank.Graph
var err error
var rtms runtime.MemStats
var counter uint

var articles_processed uint

var articles []string = []string{"Computer Science", "Information Technology", "Data Science", "Economy"}

var total, free uint64

func main() {
	articles_by_title = make(map[string]*Article)
	articles_by_offset = make(map[uint32]*string)
	var links0, links1, links2, links3, links4, links5, links6 []string
	//  articles_by_title map[string][]uint = make(map[string][]uint)
	//  articles_by_offset map[uint]string = make(map[uint]string)
	wiki_file, err = os.Open("/home/pressure679/Documents/Wiki/Pedia/enwiki-latest-pages-articles.xml")
	for num, _ := range articles {
		graph := pagerank.NewGraph()
		links0, err = append_article(articles[num])
		fmt.Println(links0)
		if err != nil { panic(err) }
		go read_mem()
		for cont {
			for links_num0, _ := range links0 {
				if _, exists := articles_by_title[links0[links_num0]]; !exists {
					links1, err = append_article(links0[links_num0])
					fmt.Println(links1)
					if err != nil { panic(err) }
					// TODO: transform this into a function[...]
					for _, root_offset := range articles_by_title[articles[num]].FileOffsets { // root[num]
						// graph.Link(uint32(articles_by_title[articles[num]]), uint32(articles_by_title[links0[links_num0]]), 1)
						for _, neighbour_offset := range articles_by_title[links0[links_num0]].FileOffsets { // links0[num0]
							graph.Link(root_offset, neighbour_offset, 1)
						}
					}
					articles_processed++
					for links_num1, _ := range links1 {
						if _, exists := articles_by_title[links1[links_num1]]; !exists {
							links2, err = append_article(links1[links_num1])
							fmt.Println(links2)
							if err != nil { panic(err) }
						}
						for _, root_offset := range articles_by_title[links0[links_num0]].FileOffsets { // links0[num0]
							for _, neighbour_offset := range articles_by_title[links1[links_num1]].FileOffsets { // links1[num1]
								graph.Link(root_offset, neighbour_offset, 1)
							}
						}
						articles_processed++
						for links_num2, _ := range links2 {
							if _, exists := articles_by_title[links1[links_num1]]; !exists {
								links3, err = append_article(links2[links_num2])
								fmt.Println(links3)
								if err != nil { panic(err) }
							}
							for _, root_offset := range articles_by_title[links1[links_num1]].FileOffsets {
								for _, neighbour_offset := range articles_by_title[links2[links_num2]].FileOffsets {
									graph.Link(root_offset, neighbour_offset, 1)
								}
							}
							articles_processed++
							for links_num3, _ := range links3 {
								if _, exists := articles_by_title[links3[links_num3]]; !exists {
									links4, err = append_article(links3[links_num3])
									fmt.Println(links4)
									if err != nil { panic(err) }
								}
								for _, root_offset := range articles_by_title[links2[links_num2]].FileOffsets {
									for _, neighbour_offset := range articles_by_title[links3[links_num3]].FileOffsets {
										graph.Link(root_offset, neighbour_offset, 1)
									}
								}
								articles_processed++
								for links_num4, _ := range links4 {
									if _, exists := articles_by_title[links4[links_num4]]; !exists {
										links5, err = append_article(links4[links_num4])
										fmt.Println(links5)
										if err != nil { panic(err) }
									}
									for _, root_offset := range articles_by_title[links3[links_num3]].FileOffsets {
										for _, neighbour_offset := range articles_by_title[links4[links_num4]].FileOffsets {
											graph.Link(root_offset, neighbour_offset, 1)
										}
									}
									articles_processed++
									for links_num5, _ := range links5 {
										if _, exists := articles_by_title[links5[links_num5]]; !exists {
											links6, err = append_article(links5[links_num5])
											fmt.Println(links6)
											if err != nil { panic(err) }
										}
										for _, root_offset := range articles_by_title[links4[links_num4]].FileOffsets {
											for _, neighbour_offset := range articles_by_title[links5[links_num5]].FileOffsets {
												graph.Link(root_offset, neighbour_offset, 1)
											}
										}
										articles_processed++
										for links_num6, _ := range links6 {
											if _, exists := articles_by_title[links6[links_num6]]; !exists {
												_, err := append_article(links6[links_num6])
												if err != nil { panic(err) }
											}
											for _, root_offset := range articles_by_title[links5[links_num5]].FileOffsets {
												for _, neighbour_offset := range articles_by_title[links6[links_num6]].FileOffsets {
													graph.Link(root_offset, neighbour_offset, 1)
												}
											}
											articles_processed++
										}
									}
								}
							}
						}
					}
				}
		}
		if cont {
		graph.Rank(0.85, 0.000001, func(node uint32, rank float64) {
		// actual[node] = rank
			// fmt.Println(articles_by_offset)
		fmt.Println(*articles_by_offset[node] + "_-_" + strconv.FormatFloat(rank, 'f', 2, 64))
		})
		} else {
			fmt.Println(articles_processed)
		}
		}
	}
}
var cont bool
func read_mem() {
	cont = true
	for {
		runtime.ReadMemStats(&rtms)
		time.Sleep(20 * time.Second)
		if rtms.Alloc / 1000000 > (total / 1000 - free / 1000) / 10 {
			cont = false
		}
	}
}
// func get_links(offset uint) (links []string, err error) {
// 	if err != nil { return nil, err }
// 	wiki_parser, err := wikiparse.NewParser(wiki_file)
// 	for num, _ := range articles {
// 		_, err = wiki_file.Seek(int64(articles_by_title[articles[num]].FileOffsets[num]), 0) // here
// 		if err != nil { return nil, err }
// 		page, err := wiki_parser.Next()
// 		if err != nil { return nil, err }
// 		if !strings.EqualFold(page.Revisions[0].Text, "") {
// 			links = wiki_parser.FindLinks(&page.Revisions[0].Text)
// 		}
// 	}
// 	return links, err
// }
func get_links(title string) (links []string, err error) {
	if err != nil { return nil, err }
	wiki_parser, err := wikiparse.NewParser(wiki_file)
	for num, _ := range articles {
		_, err = wiki_file.Seek(int64(articles_by_title[title].FileOffsets[num]), 0)
		if err != nil { return nil, err }
		page, err := wiki_parser.Next()
		if err != nil { return nil, err }
		if !strings.EqualFold(page.Revisions[0].Text, "") {
			links = wiki_parser.FindLinks(&page.Revisions[0].Text)
		}
	}
	return links, err
}
func get_offsets(title string) (offsets []uint32, err error) {
	fmt.Println(title)
	file, err := os.Open(wikipedia_index_dir + "/" + strings.ToLower(string(title[0])))
	if err != nil { panic(err) }
	bufio_reader := bufio.NewReader(file)
	for {
		line , _, err := bufio_reader.ReadLine()
		if err != nil {
			if err.Error() == "EOF" {
				break
			} else {
				return nil, err
			}
		}
		// var buffer []byte
		parts := strings.Split(string(line), "_-_")
		if strings.Contains(string(parts[0]), title) {
			if uint64_offset, err := strconv.ParseUint(parts[1], 10, 32); err == nil {
				offsets = append(offsets, uint32(uint64_offset))
			} else { return nil, err }
		}
	}
	if err = file.Close(); err != nil { return nil, err }
	return offsets, err
}
func append_article(title string) (links []string, err error) {
	Titles = append(Titles, title)
	// absLenOffsets := len(offsets)
	// absLenOffsets := len(articles_by_title[title].FileOffsets)
	offsets, err := get_offsets(Titles[len(Titles)-1])
	if err != nil { return nil, err }
	// Offsets = append(Offsets, offsets)
	links, err = get_links(title)
	if err != nil { return nil, err }
	articles_by_title[Titles[len(Titles)]].FileOffsets = offsets
	//  for offset_num, _ := range offsets {
	//  	articles_by_offset[&Offsets[len(Offsets) - len(offsets) + offset_num + 1]] = &Titles[len(Titles-1)]
	//  }
	for offset_num, _ := range offsets {
		articles_by_offset[offsets[offset_num]] = &Titles[len(Titles)-1]
	}
	return links, err
}
