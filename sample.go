// Package main provides ...
package main
import (
  "fmt"
  "os"
  "io"
	"bufio"
  "compress/bzip2"
  "strings"
	"strconv"
  "regexp"
  "sync"
  "errors"
  "github.com/dustin/go-wikiparse"
  "github.com/alixaxel/pagerank"
	"djikstra"
  //"testing"
)

type PageItems struct {
  links []string
  sections []string
  text string
	register map[string]int
  //reftohere []string
  //pagerank float64
}

func main() {
  // file := flag.String("file", "", "file to read")
  // flag.Parse()
  // var db map[string]int
	
  collection := make(map[string]*pageitems)
  file := "enwiki-latest-pages-articles1.xml-p000000010p000010000.bz2"
  wikijsonin, err := decompressbzip(file)
  /*
  file := "example.xml"
  var ior io.Reader
  osfile, err := os.Open(file)
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
  ior = bufio.NewReader(osfile)
  */
  parser, err := wikiparse.NewParser(wikijsonin)
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
  defer func() {
    if r := recover(); r != nil {
      fmt.Println("recovered in main")
    }
  }()
  var wg sync.WaitGroup
  wg.Add(1)
  go func() {
    var page *wikiparse.Page
    graph := pagerank.NewGraph()
    //defer wg.Done()
    for i := 0; i < 10; i++ {
      page, err = parser.Next()
      if err != nil {
				err = errors.New("Error while extracting wikipedia page data, attempting to recover")
				panic(err)
      }
      for i := 0; i < len(page.Revisions); i++ {
				fmt.Println(collection[page.Title])
				collection[page.Title] = &pageitems{}
				fmt.Println(collection[page.Title])
				
				// if text is not nil then add to collection text and sections to collection 
				if page.Revisions[i].Text != "" {
					collection[page.Title].text = page.Revisions[i].Text
					getsections(collection[page.Title].sections, page.Revisions[i].Text, page.Title)
				}
				
				collection[page.Title].links = wikiparse.FindLinks(page.Revisions[i].Text)
				// If there are links add them to collection
				for i := range collection[page.Title].links {
					if collection[collection[page.Title].links[i]] == nil {
						collection[collection[page.Title].links[i]] = &pageitems{}
					}
				}
      }
      // not at all optimal, we don't get node of depth 7 - first read a max of 1000 nodes with depth of max 7 in relation to our search query and find the shortest path to good related nodes (much possibly measured by pageranks).
      graph.Link(1, 7, float64(len(collection[page.Title].links)))
      graph.Rank(0.85, 0.000001, func(node int, rank float64) {
	collection[page.Title].pagerank = rank
      })
    }
    wg.Done()
  }()
  wg.Wait()
  fmt.Println(len(collection))
  var avgpr float64
  for i := range collection {
    fmt.Println(i)
    fmt.Println(collection[i].pagerank)
    /*
    fmt.Println(collection[i].links)
    fmt.Println(collection[i].reftohere)
    fmt.Println(collection[i].sections)
    */
    avgpr += collection[i].pagerank
    fmt.Println()
  }
}

// use os.Open to make an io.Reader from bzip2.NewReader(os.File) to read wikipedia xml file
func DecompressBZip (file string) (io.Reader, error) {
  osfile, err := os.Open(file)
  if err != nil {
    return nil, err
  }
  ioreader := bzip2.NewReader(osfile)
  return ioreader, nil
}

// Get sections from a wikipedia article
func GetSections(sections []string, txt, title string) error {
  re, err := regexp.Compile("[=]{2,5}.{1,50}[=]{2,5}")
  if err != nil {
    return err
  }
  if txt == "" {
    fmt.Println("page \"", title, "\" text is \"\"")
  }
  index := re.FindAllStringIndex(txt, -1)
  if len(index) == 0 {
    return errors.New("page \"" + title + "\"'s index is 0")
  }
  //fmt.Println(len(index)
  if len(index) % 2 == 0 {
    sections = make([]string, len(index) / 2)
  } else {
    sections = make([]string, (len(index) - 1) / 2)
  }
  for i := 0; i < len(index); i++ {
    if len(index) <= i * 2 + 1 { continue }
    //fmt.Println("getsections:", i, txt[index[i][0]:index[i][1]], index[i][0], index[i][1]) // debugging purposes
    sections = append(sections, txt[index[i][0]:index[i][1]])
  }
  return nil
}

// Get the byte offset of the first article with an occurence of a letter from the latin alphabet - should be used if FileExists method returns false

func Register(collection map[string]*pageitems, ioreader io.Reader) map[string]int {
  var reg map[string]int
  var count byte = 0
  alphabet := []string{"a","b","c","d","e","f","g","h","i","j","k","l","m","n","o","p","q","r","s","t","u","v","w","x","y","z","æ","ø","å"}
  indexreader := wikiparse.NewIndexReader(ioreader)
  for {
    ie, err := indexreader.Next()
    if err != nil {
      break
    }
    // if not added, add 1st character titles of wikipedia pages to our register
    if alphabet[count] != strings.ToLower(string(ie.ArticleName[0])) {
      reg[alphabet[count]] = ie.PageOffset
      count++
      continue
    }
  }
  return reg
}

// Write a database with data of the wikipedia register (data from Register(...) method) - should be used if FileExists method returns false
func CreateDB(data map[string]int) (err error) {
	f, err := os.Create("wikidb.dat")
	defer f.Close()
	if err != nil {
		log.Fatal(err)
	}
	for key, item := range(data) {
		fmt.Fprintln(f, key, item)
	}
}

// Read the database with the wikipedia register (data from Register(...) method) - should be used it FileExists method return true
func ReadDB() (data map[string]int, err error) {
	exists, err := filexists("wikidb.dat")
	if exists == false { return nil, err }
	f, err := os.Open("wikidb.dat")
	fscanner := bufio.NewScanner(f)
	strkeydataarr := make([]string, 2)
	data = make(map[string]int)
	for fscanner.Scan() {
		strkeydataarr = strings.Split(fscanner.Text(), " ")
		data[strkeydataarr[0]], err = strconv.Atoi(strkeydataarr[1])
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

// Check if a file exists, returns true if so, otherwise false, returns error for pragmatic purposes as well. - used to check if wikidb.dat exists (data from Register(...) method)
func FileExists(file string) (bool, err error) {
	_, err = os.Stat(file)
	if os.IsExist(err) { return true, nil }
	if os.IsNotExist(err) { return false, nil }
	return false, err
}

// Don't know how to apply this yet, as of now just calculate the paths from an article to another with Djikstra.
// Maybe pagerank the articles from the path and calculate shortest path to the top 5 articles with best pagerank.
// Or maybe create a DB with links with a depth of 7 from a base article, e.g Biology, History, etc. (those related to school subjects), and then take the top 5 pageranked articles from a calculated djikstra path.
func PageRank(depth, links int)  {
	
}
