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
// The Wikipedia articles can be downloaded at https://dumps.wikimedia.org/enwiki/latest/
// Package main provides ...
package main
import (
	"fmt"
	"os"
	"io"
	"io/ioutil"
	"bufio"
	"compress/bzip2"
	"strings"
	"regexp"
	"sync"
	"errors"
	"github.com/dustin/go-wikiparse"
	"github.com/pressure679/dijkstra"
	//"github.com/alixaxel/pagerank"
	"database/sql"
	"github.com/Masterminds/squirrel"
	//"testing"
)
type PageItems struct {
	Sections map[string]string // Sections (Text) from a wiki article
	Node *dijkstra.Node // Links from a wiki article
	//reftohere []string
	//pagerank float64
}
func main() {
	// var articles map[string]*PageItems
	articles := make(map[string]*PageItems)
	var wg sync.WaitGroup
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered in main")
		}
	}()
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
func ReadWikiXML(page wikiparse.Page) (articles map[string]*PageItems, err error) {
	articles = make(map[string]*PageItems)
	for i := 0; i < len(page.Revisions); i++ {
		// if text is not nil then add to articles text and sections to articles 
		if page.Revisions[i].Text != "" {
			articles[page.Title].Sections, err = GetSections(page.Revisions[i].Text, page.Title, i)
			if err != nil {
				return nil, err
			}
		}
		// Add links in article to our articles - TODO: update to make it fit with dijkstra.go
		links := wikiparse.FindLinks(page.Revisions[i].Text)
		for _, aLink := range(links) {
			articles[page.Title].Node.AppendNeighbor(aLink, 1)
		}
	}
	return articles, nil
}

// Get sections from a wikipedia article
func GetSections(page, title string, i int) (sections map[string]string, err error) {
	sections = make(map[string]string)
	// Make a regexp search object
	re, err := regexp.Compile("[=]{2,5}.{1,50}[=]{2,5}")
	if err != nil {
		return nil, err
	}

	// Check if article has a text
	// Debugging purposes
	/*
	if sections.Text == "" {
		fmt.Println("page \"", sections.Title, "\" text is \"\"")
	}
  */

	index := re.FindAllStringIndex(page, -1)
	if len(index) == 0 {
		return nil, errors.New("page's index is 0")
	}

	for i := 0; i < len(index); i++ {
		if i == 0 {
			sections["Summary"] = page[:index[i][0]-1]
		} else if i < len(index)-1 {
			sections[page[index[i][0]:index[i][1]]] = page[index[i][1]:index[i+1][0]]
		} else {
			sections[page[index[i][0]:index[i][1]]] = page[index[i][1]:len(page)]
		}
	}
	return sections, nil
}

// Create database with index of the name of first and last article of each wikipedia file.
func CreateDB() error {
	driver := "postgres"
	con, _ := sql.Open(driver, "dbname=wikidb sslmode=enable")
	cache := squirrel.NewStmtCacheProxy(con)
	files, err := ioutil.ReadDir("articles")
	if err != nil { return err }
	for _, file := range files {
		articles := make(map[string]*PageItems)
		wikijsonin, err := DecompressBZip(file)
		if err != nil { return err }
		parser, err := wikiparse.NewParser(wikijsonin)
		if err != nil { return err }
		for err == nil {
			page, err := parser.Next()
			if err != nil {
				panic(err)
			}
			articles, err = ReadWikiXML(*page)
			if err != nil {
				panic(err)
			}
		}
		for title, _ := range(articles) {
			insQ := squirrel.Insert("articles").
				Columns("article").
				Values(title).
				RunWith(db)
			insQ.Exec()
			for sectionName, body := range(articles[title].Sections) {
				insQ = squirrel.Insert(title).
					Columns(sectionName).
					Values(body).
					RunWith(db)
				insQ.Exec()
			}
		}
	}
	return nil
}

// Writing the articles items in emacs-org format (to write the path from article A to B and their top pageranking links in a presentable format)
func writeTXT(db *os.File, articles map[string]*PageItems) {
	fWriter := bufio.NewWriter()
	for articleName, _ := range(articles) {
		fmt.Fprintln(fWriter, "* " + articleName)
		fmt.Fprintln(fWriter, "** Sections")
		for sectionName, sectionText := range(articles[articleName].Sections) {
			fmt.Fprintln(fWriter, "*** " + sectionName)
			fmt.Fprintln(fWriter, "    " + sectionText)
		}
	}
	fWriter.Flush()
}

// Read the index of all wikipedia articles into a map with key as article file name and item as first and last article name
func ReadDB(articles []string) (articles map[string]*PageItems, err error) {
	articles = make(map[string]*PageItems)
	
	return articles, nil
}
