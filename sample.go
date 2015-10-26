/*
   This is basically just an API for reading a wikipedia dump from https://dumps.wikimedia.org/enwiki/,
   the search engine/database will be created with elasticsearch or bleve. - apart from go-wikiparse this has the
   GetSections method.
    Copyright (C) 2015  Vittus Mikiassen

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */
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
	"log"
	"github.com/dustin/go-wikiparse"
	//"github.com/alixaxel/pagerank"
	//"djikstra"
	//"testing"
)
type PageItems struct {
	Links []string
	Sections map[string]string
	Text string
	//reftohere []string
	//pagerank float64
}
func main() {
	var WikiArticles map[string]*PageItems
	WikiArticles = make(map[string]*PageItems)
	var wg sync.WaitGroup
	file := "enwiki-latest-pages-articles1.xml-p000000010p000010000.bz2"

	if exists, _ := FileExists(); exists == false {
		fmt.Println("file exists:", exists)
		WikiRegister = Register(WikiRegister, wikijsonin)
		CreateDB(WikiRegister)
	} else if exists == true {
		fmt.Println("file exists:", exists)
		WikiRegister, err = ReadDB()
		if err != nil {
			log.Fatal(err)
		}
	}
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered in main")
		}
	}()
	WikiArticles, err = ReadWikiXML()
	for i := range(WikiArticles) {
		fmt.Println(len(WikiArticles))
	}
	/*
	for i := range WikiArticles {
		fmt.Println(i)
		fmt.Println()
	}
	for i := range WikiRegister {
		fmt.Println(i)
		fmt.Println()
	}
  */
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

// Read Wikipedia Articles from a Wikimedia XML dump bzip file, return the Article with titles as map keys and PageItems (Links, Sections and Text) as items - Also add Section "See Also"
func ReadWikiXML() (WikiArticles map[string]*PageItems, err error) {
	wikijsonin, err := DecompressBZip(file)
	if err != nil {
		nil, err
	}
	parser, err := wikiparse.NewParser(wikijsonin)
	if err != nil {
		nil, err
	}
	for i := 0; i < 10; i++ {
		page, err := parser.Next()
		if err != nil {
			err = errors.New("Error while extracting wikipedia page data, attempting to recover")
			nil, err
		}
		WikiArticles[page.Title] = &PageItems{}
		for i := 0; i < len(page.Revisions); i++ {
			// if text is not nil then add to WikiArticles text and sections to WikiArticles 
			if page.Revisions[i].Text != "" {
				WikiArticles[page.Title].Text = page.Revisions[i].Text
				WikiArticles[page.Title].GetSections(WikiArticles[page.Title].Sections, page.Revisions[i].Text, page.Title)
			}
			
			WikiArticles[page.Title].Links = wikiparse.FindLinks(page.Revisions[i].Text)
			// If there are links add them to WikiArticles
			for i := range WikiArticles[page.Title].Links {
				if WikiArticles[WikiArticles[page.Title].Links[i]] == nil {
					WikiArticles[WikiArticles[page.Title].Links[i]] = &PageItems{} // Adds a link from the wiki article to WikiArticles
				}
			}
		}
	}
	return
}

// Get sections from a wikipedia article
func (pg PageItems) GetSections() error {
	// Make a regexp search object
	re, err := regexp.Compile("[=]{2,5}.{1,50}[=]{2,5}")
	if err != nil {
		return err
	}

	// Check if article has a text
	if pg.Text == "" {
		fmt.Println("page \"", pg.Title, "\" text is \"\"")
	}

	// 
	index := re.FindAllStringIndex(pg.Text, -1)
	if len(index) == 0 {
		return errors.New("page \"" + pg.Title + "\"'s index is 0")
	}
	pg.Sections = make(map[string]string)
	for i := 0; i < len(index); i++ {
                if i < len(index) - 1 {
                        pg.Sections[pg.Text[index[i][0]:index[i][1]]] = [pg.Text[index[i][1] + 1:index[i+1][0] - 1]]
                } else {
                        pg.Sections[pg.Text[index[i][0]:index[i][1]]] = pg.Text[index[i][1] + 1:len(pg.Text)]
                }
	}
	return nil
}

// Check if a file exists, returns true if so, otherwise false, returns error for pragmatic purposes as well. - used to check if wikidb.dat exists (data from Register(...) method)
func FileExists(file string) (bool, error) {
	_, err := os.Stat(file)
	if err == nil { return true, nil }
	if os.IsNotExist(err) { return false, nil }
	return true, err
}

// Don't know how to apply this yet, as of now just calculate the paths from an article to another with Djikstra.
// Maybe pagerank the articles from the path and calculate shortest path to the top 5 articles with best pagerank.
// Or maybe create a DB with links with a depth of 7 from a base article, e.g Biology, History, etc. (those related to school subjects), and then take the top 5 pageranked articles from a calculated djikstra path.
/*
func PageRank(depth, links int)  {
	djikstra.
}
*/

func CreateDB(WikiArticles map[string]*PageItems) error {
	var f os.File
	var lastletter string
	var err error
	for key, item = range(WikiArticles) {
		if lastletter != strings.ToLower(key[:1]) {
			lastletter = strings.ToLower(key[:1])
			switch strings.ToLower(key[:1]) {
			case "a":
				f, err = os.Create("a.dat")
				defer f.Close()
				if err != nil { return err }
			case "b":
				f, err = os.Create("b.dat")
				defer f.Close()
				if err != nil { return err }
			case "c":
				f, err = os.Create("c.dat")
				defer f.Close()
				if err != nil { return err }
			case "d":
				f, err = os.Create("d.dat")
				defer f.Close()
				if err != nil { return err }
			case "e":
				f, err = os.Create("e.dat")
				defer f.Close()
				if err != nil { return err }
			case "f":
				f, err = os.Create("f.dat")
				defer f.Close()
				if err != nil { return err }
			case "g":
				f, err = os.Create("g.dat")
				defer f.Close()
				if err != nil { return err }
			case "h":
				f, err = os.Create("h.dat")
				defer f.Close()
				if err != nil { return err }
			case "i":
				f, err = os.Create("i.dat")
				defer f.Close()
				if err != nil { return err }
			case "j":
				f, err = os.Create("j.dat")
				defer f.Close()
				if err != nil { return err }
			case "k":
				f, err = os.Create("k.dat")
				defer f.Close()
				if err != nil { return err }
			case "l":
				f, err = os.Create("l.dat")
				defer f.Close()
				if err != nil { return err }
			case "m":
				f, err = os.Create("m.dat")
				defer f.Close()
				if err != nil { return err }
			case "n":
				f, err = os.Create("n.dat")
				defer f.Close()
				if err != nil { return err }
			case "o":
				f, err = os.Create("o.dat")
				defer f.Close()
				if err != nil { return err }
			case "p":
				f, err = os.Create("p.dat")
				defer f.Close()
				if err != nil { return err }
			case "q":
				f, err = os.Create("q.dat")
				defer f.Close()
				if err != nil { return err }
			case "r":
				f, err = os.Create("r.dat")
				defer f.Close()
				if err != nil { return err }
			case "s":
				f, err = os.Create("s.dat")
				defer f.Close()
			case "t":
				f, err = os.Create("t.dat")
				defer f.Close()
				if err != nil { return err }
			case "u":
				f, err = os.Create("u.dat")
				defer f.Close()
				if err != nil { return err }
			case "v":
				f, err = os.Create("v.dat")
				defer f.Close()
				if err != nil { return err }
			case "w":
				f, err = os.Create("w.dat")
				defer f.Close()
				if err != nil { return err }
			case "x":
				f, err = os.Create("x.dat")
				defer f.Close()
				if err != nil { return err }
			case "y":
				f, err = os.Create("y.dat")
				defer f.Close()
				if err != nil { return err }
			case "z":
				f, err = os.Create("z.dat")
				defer f.Close()
				if err != nil { return err }
		fmt.Fprintln(f, key)
		for i := 0; i < len(item); i++ {
			switch i {
			case 0:
				fmt.Fprintln(f, "Links:", item)
			case 1:
				fmt.Fprintln(f, "Sections:", item)
			case 2:
				fmt.Fprintln(f, "Text:", item)
			}
		}
			}
		}
	}
}

func ReadDB(file string) (WikiArticles map[string]*PageItems, err error) {
	file, err := os.Open(file)
}
