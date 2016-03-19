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
	"mypkgs/dijkstra"
	//"github.com/alixaxel/pagerank"
	//"testing"
)
type PageItems struct {
	Sections map[string]string // Sections (Text) from a wiki article
	Node *dijkstra.Node // Links from a wiki article
	//reftohere []string
	//pagerank float64
}
func main() {
	var WikiArticles map[string]*PageItems
	WikiArticles = make(map[string]*PageItems)
	var wg sync.WaitGroup
	file := "enwiki-latest-pages-articles1.xml-p000000010p000010000.bz2"
	/*
  // Wait with this until the dijkstra method is complete
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
  */
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered in main")
		}
	}()
		wikijsonin, err := DecompressBZip(file)
	if err != nil {
		return nil, err
	}
	parser, err := wikiparse.NewParser(wikijsonin)
	if err != nil {
		return nil, err
	}
	for err == nil {
		page, err := parser.Next()
		if err != nil {
			panic(err)
		}
		WikiArticles[page.Title].ReadWikiXML(page)
		if err != nil {
			panic(err)
		}
		WikiArticles[page.Title].GetSections()
	}
	/*
	for i := range(WikiArticles) {
		fmt.Println(len(WikiArticles))
	}
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
func (collection *PageItems) ReadWikiXML(page wikiparse.Page) {
	for i := 0; i < len(collection.Revisions); i++ {
		// if text is not nil then add to WikiArticles text and sections to WikiArticles 
		if collection.Revisions[i].Text != "" {
			WikiArticles[collection.Title].GetSections(WikiArticles[collection.Title].Sections, collection.Revisions[i].Text, collection.Title)
		}

		// Add links in article to our collection
		links := wikiparse.FindLinks(collection.Revisions[i].Text)
		for _, aLink := range(links) {
			WikiArticles[collection.Title].Edges = append(WikiArticles[collection.Title].Edges, dijkstra.NewEdge(collection.Title, aLink, 1))
		}
	}
}

// Get sections from a wikipedia article
func (collection *PageItems) GetSections() error {
	// Make a regexp search object
	re, err := regexp.Compile("[=]{2,5}.{1,50}[=]{2,5}")
	if err != nil {
		return err
	}

	// Check if article has a text
	// Debugging purposes
	/*
	if collection.Text == "" {
		fmt.Println("page \"", collection.Title, "\" text is \"\"")
	}
  */

	// 
	index := re.FindAllStringIndex(collection.Text, -1)
	if len(index) == 0 {
		return errors.New("collection " + collection.Title + "'s index is 0")
	}
	
	collection.Sections = make(map[string]string)
	for i := 0; i < len(index); i++ {
		if i == 0 {
			page.Sections["Summary"] = collection.Text[:index[i][0]-1]
		} else if i < len(index)-1 {
			page.Sections[page.Text[index[i][0]:index[i][1]]] = [page.Text[index[i][1]:index[i+1][0]]]
		} else {
			page.Sections[page.Text[index[i][0]:index[i][1]]] = page.Text[index[i][1]:len(page.Text)]
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

// create and links articles (This method is first used when all articles have been read).
// The shortest path is to be calculated and returned
// The edges/links to have in memory has to be relatively low to save memory.
// The amount of edges/links is to be defined by the main method, and it should give the relevant articles as argument (by using control flow statements).
func DijkstraIt(allNodes []*dijkstra.Node, startNode, endNode *dijkstra.Node, WikiArticles map[string]*PageItems) (path []dijkstra.Path) {
	// read and copy/paste/modify from line 176 in github.com/pressure679/WikiPageRankDB/dijsktra.go
	return dijkstra.Init(allNodes, startNode, endNode)
}

// Create database with index of the name of first and last article of each wikipedia file.
func CreateDB(articles []string, wikiFileName string, WikiArticles[string]*PageItems) error {
	var firstAndLastArticle []string
	firstAndLastArticle = make([]string, 2)
	counter := 0
	mapLen := len(articles)
	for key, _ = range(articles) {
		counter++
		switch counter {
		case 1:
			firstAndLastArticle[0] = key + "$" + 
		case mapLen:
			firstAndLastArticle[1] = key
		}
	}
	if exists, existserr := FileExists(firstAndLastArticle[0] + "-" + firstAndLastArticle[1] + ".org"); existserr == nil {
		db, err := os.OpenFile(firstAndLastArticle[0] + "-" + firstAndLastArticle[1] + ".org", os.O_RDWR|os.O_APPEND, 0660)
		defer db.Close()
		if err != nil {
			return err
		}
		writeTXT(db, WikiArticles)
	} else if exists == false {
		db, err := os.Create(firstAndLastArticle[0] + "-" + firstAndLastArticle[1] + ".org")
		defer db.Close()
		if err != nil {
			return err
		}
		writeTXT(db, WikiArticles)
	}
}

// Writing the WikiArticles items in emacs-org format (used by CreateDB)
func writeTXT(db os.File, articles WikiArticles[string]*PageItems) {
	fwriter := bufio.NewWriter(db)
	for articleName, _ = range(articles) {
		fmt.Fprintln(fwriter, "* " + articleName)
		fmt.Fprintln(fwriter, "** Sections")
		for sectionName, sectionText := range(WikiArticles[articleName].Sections) {
			fmt.Fprintln(fwriter, "*** " + sectionName)
			fmt.Fprintln(fwriter, "    " + sectionText)
		}
		fmt.Fprintln(fwriter, "** Links")
		// TODO: differ between Edges and Nodes (write the node number of this article, and then the node number for the linked article's - also write 1st node number and last node number in the file's name.
		for linkNum, linkName := range(WikiArticles[key].Edges) {
			fmt.Fprintln(fwriter, "   - " + linkNum + " - " + linkName)
		}
		// TODO: add a section with links to articles with a depth of 7 (dijkstra + pagerank noticeable optimization)
	}
	fwriter.Flush()
}

// Read the index of all wikipedia articles into a map with key as article file name and item as first and last article name
func (page *PageItems) ReadDB(file string, collection map[string]*PageItems) (collection map[string]*PageItems) error {
	// wikiArticles = make(map[string]string)
	var splittedIndex []string = make([]string, 3)
	var splittedDijkstra []string = make([]string, 2)
	var index, section, link bool = false
	var buffer string
	if err := FileExists("db.dat"); err == nil {
		file, err := os.Open("db.dat")
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			// TODO: read links and sections, links; do strings.Split when reading a link org-section, then append to edge. and read 
			switch {
			case strings.EqualFold(scanner.Text()[0:2], "* "):
				collection[scanner.Text()[2:]] = &PageItems{}
				if section == true { section = false }
				if link == true { link = false }
			case strings.EqualFold(scanner.Text()[0, 3], "** "):
				buffer = scanner.Text()[3:]
				collection[buffer] = &PageItems{}
				if strings.EqualFold(buffer, "Links"} {
					section = false
					link = true
				}
				if strings.EqualFold(buffer, "Sections") {
					link = false
					section = true
				}
			case strings.EqualFold(scanner.Text()[0, 4], "*** "):
				buffer = strings.Split(scanner.Text(), " ")[4:]
			}
			if link == true && !strings.EqualFold(buffer, "Links") {
				
			}
			if section == true && !strings.EqualFold(buffer, "Sections") {
				wikiArticles[buffer] = splitted[1] + "-" + splitted[2]
			}
		}
	} else if exists == false {
		return errors.New("Database not created")
	}
	return nil
}
