// Package main provides ...

package main

import (
  "fmt"
  "os"
  "io"
  "compress/bzip2"
  "regexp"
  "strings"
  "sync"
  "testing"
  "errors"
  "github.com/dustin/go-wikiparse"
  //"github.com/alixaxel/pagerank"
)

type pageitems struct {
  subtitles []string
  links []string
  text string
  offset int64
  reftohere []string
  pagerank float64
}

type alphabetpos int

func main() {
  // file := flag.String("file", "", "file to read")
  // flag.Parse()
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
  var page *wikiparse.Page
  collection := make(map[string]*pageitems)
  defer func() {
    if r := recover(); r != nil {
      fmt.Println("recovered in main")
    }
  }()
  var wg sync.WaitGroup
  wg.Add(1)
  go func(*testing.B) {
    defer wg.Done()
    for i := 0; i < 10; i++ {
      page, err = parser.Next()
      if err != nil {
	err = errors.New("Error while extracting wikipedia page data, attempting to recover")
	panic(err)
      }
      //fmt.Println(page.Revisions)
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
    }
  }()
  go func(*testing.B) {
    defer wg.Done()
    var source, target int
    var weight float64
    for i := range collection {
      source = 1
      target = 10
      weight = len(collection[i].links)
      collection[i].pagerank = 

  }
  wg.Wait()
  fmt.Println(len(collection))
  for i := range collection {
    fmt.Println(i)
    fmt.Println(collection[i].links)
    fmt.Println(collection[i].reftohere)
    fmt.Println(collection[i].subtitles)
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
func checkcharpos(char string) int {
  alphabet = []string{"a","b","c","d","e","f","g","h","i","j","k","l","m","n","o","p","q","r","s","t","u","v","w","x","y","z","æ","ø","å"}
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
  var i int
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
