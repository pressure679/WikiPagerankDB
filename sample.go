// Package main provides ...

package main

import (
  "fmt"
  "os"
  "io"
  "compress/bzip2"
  "regexp"
  "errors"
  "github.com/dustin/go-wikiparse"
  "sync"
  //"testing"
  "strings"
  "github.com/alixaxel/pagerank"
)

type pageitems struct {
  alphabetoffset int64
  links []string
  reftohere []string
  subtitles []string
  text string
  pagerank float64
}

func main() {
  // file := flag.String("file", "", "file to read")
  // flag.Parse()
  //var db map[string]int
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
	// if text is not nil then add to collection text and subtitles to collection 
	fmt.Println(collection[page.Title])
	collection[page.Title] = &pageitems{}
	fmt.Println(collection[page.Title])
	if page.Revisions[i].Text != "" {
	  collection[page.Title].text = page.Revisions[i].Text
	  getsubtitles(collection[page.Title].subtitles, page.Revisions[i].Text, page.Title)
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
    fmt.Println(collection[i].subtitles)
    */
    avgpr += collection[i].pagerank
    fmt.Println()
  }
}

func decompressbzip(file string) (io.Reader, error) {
  osfile, err := os.Open(file)
  if err != nil {
    return nil, err
  }
  ioreader := bzip2.NewReader(osfile)
  return ioreader, nil
}

func getsubtitles(subtitles []string, txt, title string) error {
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
    subtitles = make([]string, len(index) / 2)
  } else {
    subtitles = make([]string, (len(index) - 1) / 2)
  }
  for i := 0; i < len(index); i++ {
    if len(index) <= i * 2 + 1 { continue }
    //fmt.Println("getsubtitles:", i, txt[index[i][0]:index[i][1]], index[i][0], index[i][1]) // debugging purposes
    subtitles = append(subtitles, txt[index[i][0]:index[i][1]])
  }
  return nil
}

func register(collection map[string]*pageitems, ioreader io.Reader) map[string]int {
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
/*
func checkcharpos(char string) int {
  const (
    a alphabetpos = 0 + iota
    b
    c
    d
    e
    f
    g
    h
    i
    j
    k
    l
    m
    n
    o
    p
    q
    r
    s
    t
    u
    v
    w
    x
    y
    z
    æ
    ø
    å
  )
  for i := 0; i < len(alphabet); i++ {
    if char == alphabet[i] {
      break
    }
  }
  switch (i) {
  case a:
    return a
  case b:
    return b
  case c:
    return c
  case d:
    return d
  case e:
    return e
  case f:
    return f
  case g:
    return g
  case h:
    return h
  case i:
    return i
  case j:
    return j
  case k:
    return k
  case l:
    return l
  case m:
    return m
  case n:
    return n
  case o:
    return o
  case p:
    return p
  case q:
    return q
  case r:
    return r
  case s:
    return s
  case t:
    return t
  case u:
    return u
  case v:
    return v
  case w:
    return w
  case x:
    return x
  case y:
    return y
  case z:
    return z
  case æ:
    return æ
  case ø:
    return ø
  case å:
    return å
  }
}
*/
