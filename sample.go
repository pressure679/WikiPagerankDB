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

// TODO: Big picture from here is to make sorting algorithm for pagerank function, update SqlInsert, then the "installation/CreateDB" part of the program is finished.
// Then prettify the output of WriteTXT with a CSS just for the sake of it.
// Then when the program is to utilize the dijkstra's algorithm and have found a path, in the main function, use the pagerank part of the DB to load all articles with a max distance/depth of 7 from the MySQL DB (test it to utilize RAM and/or CPU best, e.g how many articles to load at a time), then discard the neighboring articles from the RAM if they do not have the base article as a neighbor.
// Then get the top 5 or 10 best pageranking neighboring articles from each article in the path (len(path) / 7 * 3, then 5, then 7 if RAM is clocked), then use WriteTXT function to write the summaries of each article (path and top 5 or 10 pageranking algorithms).

package main
import (
	"fmt"
	"os"
	"io"
	"io/ioutil"
	"bufio"
	"compress/bzip2"
	"strings"
	"strconv"
	"regexp"
	// "sync"
	"errors"
	"github.com/dustin/go-wikiparse"
	"github.com/pressure679/dijkstra"
	//"github.com/alixaxel/pagerank"
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	"github.com/boltdb/bolt"
	//"testing"
)

// Data is pagerank, key is depth from base article, item is pagerank.
type Data map[uint8]float64
type PageItems struct {
	// Sections, title indicated by key, item is content/text
	Sections map[string]string
	
	// The NodeID, used for dijkstra's algorithm
	NodeID uint32
	
	// This article's pagerank for another article (the other articles has a max depth of 7 indicated by the 2nd map's key, pagerank is indicated by 2nd map's item)
	Pagerank map[string]Data
	
	// links from this article, used to collect them for the MySQL DB, after that the program will use them to utilize the DB for Dijkstra's algorithm and the Pagerank algorithm.
	Links []string
}

func main() {
	// TODO: Utilize the functions, get the wikimedia xml dumps and clean the code for bugs.
	
	// TODO: add concurrency to each function if needed and also to function calls.
}

// Read all Bzipped Wikimedia XML files from "articles" dir.
func GetFilesFromArticlesDir() (files []string, err error) {
	osFileInfo, err := ioutil.ReadDir("articles")
	if err != nil { return nil, err }
	for _, fileInfo := range osFileInfo {
		if !fileInfo.IsDir() {
			files = append(files, fileInfo.Name())
		}
	}
	return files, nil
}

// uses os.Open to make an io.Reader from bzip2.NewReader(os.File) to read wikipedia xml file
func DecompressBZip (file string) (io.Reader, error) {
	osfile, err := os.Open(file)
	if err != nil {	return nil, err	}
	ioreader := bzip2.NewReader(osfile)
	return ioreader, nil
}

// Reads Wikipedia articles from a Wikimedia XML dump bzip file, return the Article with titles as map keys and PageItems (Links, Sections and Text) as items - Also add Section "See Also"
func ReadWikiXML(page wikiparse.Page) (articles map[string]PageItems, err error) {
	articles = make(map[string]PageItems)
	var tmp PageItems
	for i := 0; i < len(page.Revisions); i++ {
		tmp = articles[page.Title]
		// if text is not nil then add to articles text and sections to articles 
		if page.Revisions[i].Text != "" {
			// tmp[page.Title].Sections = make(map[string]string)
			tmp.Sections, err = GetSections(page.Revisions[i].Text, page.Title, i)
			if err != nil {
				return nil, err
			}
		}
		links := wikiparse.FindLinks(page.Revisions[i].Text)
		for _, link := range(links) {
			tmp.Links = append(tmp.Links, link)
		}
		articles[page.Title] = tmp
	}
	return articles, nil
}

// Gets sections from a wikipedia article
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

// Pageranks articles (map's item) with a given depth/distance (map's key) from a base article which must have a max distance of 7.
// TODO: Sort the return value by the highest pagerank (2nd map's item) (maybe in main function)
func Pagerank(articles map[uint8][]string) (pagerank map[string]Data) {
	graph := pagerank.NewGraph()

	var x uint16 = 0

	for i := 0; i < len(articles); i++ {
		for key, item := range(articles[i]) {
			graph.Link(1, key, i)
		}
	}
	
	x = 0
	var xx uint16 = 0
	pagerank = make(map[string]Data)
	graph.Rank(0.85 /*put damping factor here or just settle with weighing the graph?*/, 0.000001 /*precision*/, func(node uint64, rank float64) {
		pagerank[articles[x][xx]][x] = rank
		if xx == len(articles[x]) {
			x++
			xx = 0
		}
		xx++
	})
}

func (Page PageItems) BoltCreate(file string) (err error) {
	db, err := bolt.Open("articles/" + file, 0666, nil)
	if err != nil { return err }
	return nil
}

// Insert one wikimedia dump file at a time, the db file is named "[a-z]-[az]*", the regexp's are the letter of the 1st and last article in the dump file (get that in the main func)
func (Page PageItems) BoltInsert(articles map[string]PageItems) (err error) {
	// range over files created by BoltCreate (the names are named by the range of articles they contain), then insert the articles obtained from ReadWikiXML, articles items are got by GetSections and PageRank functions).
	var files []string
	osFileInfo, err := ioutil.ReadDir("articles")
	if err != nil { return err }
	// range over wikipedia 
	for _, fileInfo := range osFileInfo {
		if !fileInfo.IsDir() {
			files = append(files, fileInfo.Name())
		}
	}
	for key, _ := range articles {
		b, err := tx.CreateBucket([]byte(key))
		if err != nil { return err }
			// Puts the name of the articles sections and section content into the bolt bucket, so the TX's bucket name is the article name, and TX's bucket content's key is the article's sections, pagerank, links and so forth (see type article struct for full content).
		for sectionKey, sectionText := range articles[key].Sections {
			if err := b.Put([]byte(sectionKey), []byte(sectionText)); err != nil { return err }
		}
		// Puts the pagerank into the bucket, so the //TODO: Make comments
		for prKey, prData := range items.Pagerank {
			for depth, pagerank := range prData {
				if err := b.Put([]byte("pr_target-" + prKey), []byte(strconv.Itoa(depth) + "-" + strconv.Itoa(pagerank))); err != nil { return err }
			}
		}
		for _, link := range items.Links {
			if err := b.Put([]byte("Links"), []byte(items.Links)); err != nil { return err }
		}
		if err := b.Put([]byte("NodeID"), []byte(items.NodeID)); err != nil { return err }
	}
	return nil
}

func (Page PageItems) BoltGet(tx *bolt.Tx, articles, keys []string) (articlesData map[string]PageItems) {
	var tmp PageItems
	articlesData = make(map[string]PageItems)
	var counter uint = -1
	// range over articles requested and get their neighboring articles with a max depth of 7.
	for _, article := range articles {
		// Get the value (the strings in the "keys" string array); the Data keys, e.g Pagerank, NodeID, article name, sections etc..
		for _, value := range keys {
			// Following if/else if should really be a switch, but we check if the key's value is NodeID, Pagerank and so forth.
			// Check BoltInsert to see how the keys are made.
			if strings.EqualFold(key, "NodeID") {
				tmp.NodeID = strconv.Atoi(string(value))
			} else if strings.Contains(key, "pr_target-") {
				tmp = PageItems{
					Pagerank: map[string]Data{
						article: map[uint8]float64{},
					},
				}
				pr_data := strings.Split(value, "-")
				tmp.Pagerank[article] = Data{}
				// TODO: Add articles, check how the db delegates keys and values of pageranks.
				tmp.Pagerank[article].Data = make(map[uint8]float64)
			}
			value := tx.Bucket([]byte(article)).Get([]byte(key))
		}
	}
	return value
}
func (Page PageItems) BoltUpdate(tx *bolt.Tx, article map[string]PageItems) (err error) {
	for key, items := range article {
		b, err := tx.Bucket([]byte(key))
		for sectionKey, sectionText := range items.Sections {
			if err := b.Put([]byte(sectionKey), []byte(sectionText)); err != nil { return err }
		}
	}
}

// Dijkstra's algorithm, used to find shortest path (if any) between 2 articles.
func Dijkstra(request dijkstra.Request) (path []string) {
	return dijkstra.Get(request)
}

// TODO: optimize the emacs-org format text written in WriteTXT function to be pretty (just make the text pretty for when the system-call to format the org-txt to html format, maybe just make a CSS for it).
// Also when the articles, separately, has to be written, then make it like "1 - <article title>", "2 - <article title>" with the number relatively to the path number from article A to B. The neighboring articles to the base articles are written in the folders with folder name given by the path number. The neighboring article's name are numbered by their distance then their pagerank and then their title.
// Writes the articles items in emacs-org format (to write the path from article A to B and their top pageranking links in a presentable format)
func WriteTXT(articles map[string]PageItems) (err error) {
	// fWriter := bufio.NewWriter(ioWriter)
	for articleName, _ := range(articles) {
		file, err := os.Create(articleName)
		defer file.Close()
		file.WriteString("* " + articleName)
		file.WriteString("** Sections")
		for sectionName, sectionText := range(articles[articleName].Sections) {
			file.WriteString("*** " + sectionName)
			file.WriteString("    " + sectionText)
		}
	}
}

// TODO: make a basic chatbot which uses Eliza's principle of chatting, the basic outrule for chatting uses Bloom's Taxonomical levels to present data/chat with a user.
// The packages to be used here are word2sentence etc., elasticsearch, and it should utilize the information written to a file in emacs-org format/html-format.
// From here the chatbot is to be written, just to make it J.A.R.V.I.S-like. the chatbot will use ELIZA like chat functionality, e.g explain how one article relates to another by using Bloom's Taxonomical levels (here word2sentence etc. and Elasticsearch will be used).
// The chatbot will load the summaries of the written (from WriteTXT) articles that are in the path of article A, B, C... and their top 5 or 10 pageranking articles.
// The chatbot may start with "Hi, which topics would you like me to chat with you about?". And the chatbot may want to only use a max of 4-7 new words in each message to not strain the short-term memory. It's functionality will be different than the product of WriteTXT function; it will connect the paths between the topics, but the pagerank functionality will work somehow indifferable; IMPORTANT: it will sum up pageranks if a neighboring article has 2 base articles as neighbors. (should maybe also be used for the WriteTXT function).
// Create a neural network for the chatbot with a light prediction analysis algorithm (prediction algorithm trained by user's behaviour) to make it smarter.. Maybe the neural network is not needed since Dijkstra and Pagerank makes up for it; the shortest part and the top pageranking path. Maybe add a neural network to find the path from shortest path and top pageranking path to make it better to chat with humans rather than machines(?).
// Here I may also add a xml dump file of articles from HowStuffWorks if it provides such to add a functinality to the chatbot to answer "how" questions.

// TODO: if the chatbot functionality works well; hook it up to 1 or more irc channels.

// TODO: make a GUI for Android for the chatbot functionality
