/*
	 This is basically just an API for reading a wikipedia dump from https://dumps.wikimedia.org/enwiki/,
	 the search engine/database will be created with elasticsearch or bleve. - apart from go-wikiparse this has the GetSections method.
		Copyright (C) 2015-2016 Vittus Mikiassen

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
	"compress/bzip2"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	// "bufio"
	"regexp"
	"sort"
	"strconv"
	"strings"
	// "runtime"
	// "flag"
	// "errors"
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	"github.com/dustin/go-wikiparse"
	// "github.com/arnauddri/algorithms/data-structures/graph"
	"github.com/alixaxel/pagerank"
	"rosettacode/dijkstra" // See RosettaCode, I cannot take credit for this package.
	// "sync"
	// "testing"
)

type SuccessorPr map[string][]byte
type PageItems struct {
	// Sections, title indicated by key, item is content/text
	Sections map[string]string

	// The NodeID, used for grahping/treeing
	NodeID []byte

	// Weight/distance from root
	Weight [6]bool

	// Links from this article, used to collect them for the MySQL Db, after that the program will use them to utilize the Db for Dijkstra's algorithm and the Pagerank algorithm.
	Links string

	Pageranks map[uint8]SuccessorPr
}
type WikiGraph struct {
	Vertices map[string]map[[6]bool][]string
}
type MySQL struct {
	// For inserting articles
	LinksLen int
	SqlLinksDT string
	NodeIDLen int
	SqlNodeIDDT string
	SectionLen int
	SqlAbsSectionDT string

	// For selecting articles for the graph
	Graph WikiGraph
}

func main() {
	// createDb := flag.String("create-db", "", "Whether to create db - It has to be to start with, otherwise error will occur.")
	// updateDb := flag.String("update-db", "", "Whether or not to update db")
	// base := flag.Bool("base", false, "A base article, the article used to communicate with other users.")
	// link := flag.Bool("link", false, "A target article, an article and etc. to link by either getting the Bloom-circle from another user or to create it your own pc, after the Bloom-circle is created it will link the your base articles with the new.")
	// flag.Parse()

	defer recover()
	dumpFiles, err := GetFilesFromArticlesDir()
	if err != nil {
		panic(err)
	}
	// fmt.Println(dumpFiles)
  db, err := sql.Open("mysql", "naamik:WM\"slLE/vm.R@tcp(localhost:3306)/example_db")
	for _, file := range dumpFiles {
		fmt.Println("Reading", file)
		// fmt.Println(file)
		ioReader, allocSize, err := DecompressBZip(file)
		if err != nil { fmt.Println("ioReader, DecompressBZip error"); panic(err) }
		wikiParser, err := wikiparse.NewParser(ioReader)
		if err != nil { fmt.Println("Error occured"); panic(err) }
		// fmt.Println("Read", file)
		for {
			articles := make(map[string]PageItems)
			page, err := wikiParser.Next()
			if err != nil {
				if strings.EqualFold(err.Error(), io.EOF) { fmt.Println("Wiki dump", file, "read"); break } else { fmt.Println("Error occured:", err); continue } /* panic(err) */
			}
			articles[page.Title], err = ReadWikiXML(*page)
			if err != nil { fmt.Println("Error occured:", err.Error(), "\n" + page.Title, "wasn't read"); break }
			err = SqlInsertArticle(db, articles)
			if err != nil { panic(err) }
			articles = nil
		}
		// go runtime.GC()
		fmt.Println("Appended", file)
	}

	/* if *base {
	baseFile, err := os.Open("bases.txt")
	if err != nil { panic(err) }
	fileReader := bufio.NewReader(baseFile)
	for err == nil {
		baseArticle, _, err := fileReader.ReadLine()
		if err != nil { panic(err) }
		// wikiGraph, err := BoltGetGraph(string(baseArticle))
		if err != nil { panic(err) }
		articles, err := BoltGetArticles(string(baseArticle))
		if err != nil { panic(err) }
		articles, err = BoltGetChildren(string(baseArticle), articles)
		if err != nil { panic(err) }
		articlesPr, err := PagerankGraph(string(baseArticle), articles[string(baseArticle)].Pageranks)
		if err != nil { panic(err) }
		for depth, item := range articlesPr {
			articles[string(baseArticle)].Pageranks[depth] = item
		}
		if err := BoltInsertPagerank(articles); err != nil { panic(err) }
		err = WriteTxt(articles)
		if err != nil { panic(err) }
	} */
	// Then use the "link" flag to execute the dijkstra function from each created base from earlier. The top ranking shared pages with a max distance of 7 from bases should be returned as well.
	// Use the WriteTxt function for
}

// Read all names of Bzipped Wikimedia XML files from "articles" dir.
func GetFilesFromArticlesDir() (files []string, err error) {
	osFileInfo, err := ioutil.ReadDir("articles")
	if err != nil {
		return nil, err
	}
	for _, fileInfo := range osFileInfo {
		if !fileInfo.IsDir() {
			files = append(files, fileInfo.Name())
		}
	}
	return files, nil
}

// uses os.Open to make an io.Reader from bzip2.NewReader(os.File) to read wikipedia xml file
func DecompressBZip(file string) (ioReader io.Reader, fileSize int64, err error) {
	osFile, err := os.Open("D:/Documents/articles/" + file)
	if err != nil {
		return nil, -1, err
	}
	fileStat, err := osFile.Stat()
	if err != nil {
		return nil, fileStat.Size(), err
	}
	ioReader = bzip2.NewReader(osFile)
	return ioReader, fileStat.Size(), nil
}

// Reads Wikipedia articles from a Wikimedia XML dump bzip file, return the Article with titles as map keys and PageItems (Links, Sections and Text) as items - Also add Section "See Also"
func ReadWikiXML(page wikiparse.Page) (pageItems PageItems, err error) {
	for i := 0; i < len(page.Revisions); i++ {
		// if text is not nil then add to articles text and sections to articles
		if page.Revisions[i].Text != "" {
			pageItems.Sections = make(map[string]string)
			pageItems.Sections, err = GetSections(page.Revisions[i].Text, page.Title)
			if err != nil { return nil, err }
			Links := wikiparse.FindLinks(page.Revisions[i].Text)
			for num, link := range Links {
				if num == 0 {
					pageItems.Links = append(pageItems.Links, link)
				} else {
					pageItems.Links = append(pageItems.Links, "-" + link)
				}
			}
			pageItems.NodeID = []byte(strconv.Itoa(int(page.Revisions[i].ID)))
		}
	}
	return pageItems, nil
}

// Gets sections from a wikipedia article, page is article content, title is article title
func GetSections(page, title string) (sections map[string]string, err error) {
	sections = make(map[string]string)
	// Make a regexp search object for section titles
	re, err := regexp.Compile("[=]{2,5}.{1,50}[=]{2,5}")
	if err != nil {
		return nil, err
	}
	// fmt.Println(title + "\n" + page)
	index := re.FindAllStringIndex(page, -1)
	// return nil, errors.New("page's index is 0") ?
	// if len(index) == 0 { return nil, errors.New("page's index is 0") }
	// Look at how regex exactly does this
	/* if strings.EqualFold(strings.ToLower(title), "southern hemisphere" ) {
		fmt.Println(index, len(page))
		fmt.Println(page)
	} */
	for i := 0; i < len(index)-1; i++ {
		// fmt.Println(len(page), index[i])
		if index[i] != nil {
			if i == 0 {
				// sections["Summary"] = page[:index[i + 1][0]-1]
				/* fmt.Println(index)
				fmt.Println(page[:index[i][0]-1])
				fmt.Println(page) */
				sections["Summary"] = page[:index[i][1]-1] // Error: slice out of bounds at article "Southern Hemisphere"
			} else if i < len(index)-1 {
				sections[page[index[i][0]:index[i][1]]] = page[index[i][1]:index[i+1][0]] // Assume this will create an error like the "if i == 0" condition error (maybe not)
				// sections[page[index[i][0]:index[i]]] = page[index[i]:index[i+1]]
			} else {
				sections[page[index[i][0]:index[i][1]]] = page[index[i][1]:len(page)] // Assume this will create an error like the "if i == 0" condition error /maybe not)
				// sections[page[index[i]:index[i]]] = page[index[i]:len(page)]
			}
		}
	}
	return sections, nil
}

func (self MySQL) InsertArticle(db sql.DB, articles map[string]PageItems) (err error) {
	for title, items := range articles {
		self.LinksLen = len(items.Links)
		// Determine length of the string containing links and assign mysql datatype accordingly
		switch {
		case self.LinksLen < 256:
			SqlLinksDT = "TINYBLOB"
		case len(links) >= 256 && self.LinksLen < 65536:
			SqlLinksDT = "BLOB"
		case self.LinksLen >= 65536 && self.LinksLen < 16777216:
			SqlLinksDT = "MEDIUMBLOB"
		case self.LinksLen >= 16777216 && self.LinksLen < 4294967296:
			SqlLinksDT = "LONGBLOB"
		}
		// Determine size of the int containing NodeID/ID and assign mysql datatype accordingly
		switch {
		case items.NodeID < 256:
			self.SqlNodeIDDT = "TINYINT"
		case items.NodeID >= 256 && items.NodeID < 65536:
			self.SqlNodeIDDT = "INT"
		case items.NodeID >= 65536 && items.NodeID < 16777216:
			SqlNODEIDDT = "MEDIUMINT"
		case items.NodeID >= 16777216 && items.NodeID < 4294967296:
			SqlNODEIDDT = "BIGINT"
		}
		_, err = db.Exec("CREATE TABLE " +
			title +
			" (NodeID " +
			self.SqlNodeIDDT +
			", Links " +
			self.SqlLinksDDT +
			";")
		_, err = db.Exec("INSERT INTO " +
			title +
			" (NodeID, Links) VALUES (?, ?);",
			items.NodeID, items.Links)
		if err != nil { return err }
		// TODO: Update to fit with CREATE TABLE
		for sectionTitle, sectionBody := range items.Sections {
			self.SectionLen = len(sectionBody)
			// Determine length of the string containing section and assign mysql datatype accordingly
			switch {
			case self.SectionLen < 256:
				self.SqlAbsSectionDT = "TINYBLOB"
			case self.SectionLen >= 256 && self.SectionLen < 65536:
				self.SqlAbsSectionDT = "BLOB"
			case self.SectionLen >= 65536 && self.SectionLen < 16777216:
				self.SqlAbsSectionDT = "MEDIUMBLOB"
			case self.SectionLen >= 65536 && self.SectionLen < 4294967296:
				self.SqlAbsSectionDT = "LONGBLOB"
			}
			_, err = db.Exec("ALTER TABLE " +
				title +
				" ADD COLUMN " +
				sectionTitle +
				" " +
				self.SqlAbsSectionDT +
				";",
			)
			if err != nil { return err }
			_, err = db.Exec("INSERT INTO " +
				title +
				"(" +
				sectionTitle +
				") VALUES (?);",
				sectionBody)
			if err != nil { return err }
		}
	}
	return nil
}
func (self MySQL) SqlGetGraph(db sql.DB, article string) {
	var Link0 []string
	var Link1 []string
	var Link2 []string
	var Link3 []string
	var Link4 []string
	var Link5 []string
	var Link6 []string
	rows0, err := db.Query("SELECT Links FROM " + article);
	if err != nil { return nil, err }
	self.Graph.Vertices[article] = make(map[[6]bool]string)
	for rows0.Next() {
		err = rows0.Scan(&Links0)
		for _, link0 := range Links0 {
			self.Graph.Vertices[article][[6]bool{true, false, false, false, false, false, false}] = append(self.Graph.Vertices[article][[6]bool{true, false, false, false, false, false, false}], link0)
		}
		
		rows1, err := db.Query("SELECT Links FROM " + link0);
		if err != nil { return nil, err }
		self.Graph.Vertices[article] = make(map[[6]bool]string)
		for rows1.Next() {
			err = rows1.Scan(&Links1)
			for _, link1 := range Links1 {
				self.Graph.Vertices[article][[6]bool{false, true, false, false, false, false, false}] = append(self.Graph.Vertices[article][[6]bool{false, true, false, false, false, false, false}], link1)
			}
			
			rows2, err := db.Query("SELECT Links FROM " + link1);
			if err != nil { return nil, err }
			self.Graph.Vertices[article] = make(map[[6]bool]string)
			for rows2.Next() {
				err = rows2.Scan(&Links2)
				for _, link2 := range Links2 {
					self.Graph.Vertices[article][[6]bool{false, false, true, false, false, false, false}] = append(self.Graph.Vertices[article][[6]bool{false, false, true, false, false, false, false}], link2)
				}
				
				rows3, err := db.Query("SELECT Links FROM " + link2);
				if err != nil { return nil, err }
				self.Graph.Vertices[article] = make(map[[6]bool]string)
				for rows3.Next() {
					err = rows3.Scan(&Links3)
					for _, link3 := range Links3 {
						self.Graph.Vertices[article][[6]bool{false, false, false, true, false, false, false}] = append(self.Graph.Vertices[article][[6]bool{false, false, false, true, false, false, false}], link3)
					}
					
					rows4, err := db.Query("SELECT Links FROM " + link3);
					if err != nil { return nil, err }
					self.Graph.Vertices[article] = make(map[[6]bool]string)
					for rows4.Next() {
						err = rows4.Scan(&Links4)
						for _, link4 := range Links4 {
							self.Graph.Vertices[article][[6]bool{false, false, false, false, true, false, false}] = append(self.Graph.Vertices[article][[6]bool{false, false, false, false, true, false, false}], link4)
						}
						
						rows5, err := db.Query("SELECT Links FROM " + link4);
						if err != nil { return nil, err }
						self.Graph.Vertices[article] = make(map[[6]bool]string)
						for rows4.Next() {
							err = rows4.Scan(&Links4)
							for _, link4 := range Links4 {
								self.Graph.Vertices[article][[6]bool{false, false, false, false, true, false}] = append(self.Graph.Vertices[article][[6]bool{false, false, false, false, true, false}], link4)
							}
							
							rows6, err := db.Query("SELECT Links FROM " + link5);
							if err != nil { return nil, err }
							self.Graph.Vertices[article] = make(map[[6]bool]string)
							for rows6.Next() {
								err = rows6.Scan(&Links6)
								for _, link6 := range Links6 {
									self.Graph.Vertices[article][[6]bool{false, false, false, false, false, false, true}] = append(self.Graph.Vertices[article][[6]bool{false, false, false, false, false, false, true}], link6)
								}
							}
						}
					}
				}
			}
		}
	}
	return articles, nil
}
// TODO: Finish this function.
func (self MySQL) SqlGetArticles(db sql.DB, articles []string) (articles map[string]PageItems, err error) {
	var tmp PageItems
	for _, title := range articles {
		rows, err := db.Query("SELECT * FROM " + title);
		if err != nil { return nil, err }
	}
	return articles, nil
}

// TODO: use arnauddris algorithm pkg to make a directed graph of wikipedia with a max edge count from a root vertex of 7, then utilize the PagerankGraph function
func PagerankGraph(title string, children map[uint8]SuccessorPr) (map[uint8]SuccessorPr, error) {
	pr := make(map[uint8]SuccessorPr)
	pagerankGraph.Rank(0.85 /*put damping factor here or just settle with weighing the graph?*/, 0.000001 /*precision*/, func(node uint32, rank float64) {
		bufferRank := []byte(strconv.FormatFloat(rank, 10, 6, 64))
		pr[articlesDepth[articlesTitle[strconv.Itoa(int(node))]]] = SuccessorPr{articlesTitle[strconv.Itoa(int(node))]: bufferRank}
	})
	return pr, nil
}

// graphs the articles, preferred input is a graph with 7 a distance/number of node generations of 7.
// TODO
func Dijkstra(request dijkstra.Request) (path []string) {
	return dijkstra.Get(request)
}

type Pagerank float64
type SortedPageranks []Pagerank

func (a SortedPageranks) Len() int           { return len(a) }
func (a SortedPageranks) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SortedPageranks) Less(i, j int) bool { return a[i] < a[j] }
func WriteTxt(articles map[string]PageItems) (err error) {
	var pageranks map[string][]Pagerank
	// fWriter := bufio.NewWriter(ioWriter)
	for articleName, _ := range articles {
		file, err := os.Create(articleName + ".org")
		if err != nil {
			return err
		}
		defer file.Close()
		indexFile, err := os.Create("index-" + articleName + ".org")
		if err != nil {
			return err
		}
		defer indexFile.Close()
		file.WriteString("* " + articleName)
		file.WriteString("** Sections")
		for sectionName, sectionText := range articles[articleName].Sections {
			file.WriteString("*** " + sectionName)
			file.WriteString("    " + sectionText)
		}
		for depth, _ := range articles[articleName].Pageranks {
			indexFile.WriteString("* " + strconv.Itoa(int(depth)))
			pagerankIndex := make(map[float64]string)
			for article, pagerank := range articles[articleName].Pageranks[depth] {
				float64Pr, err := strconv.ParseFloat(string(pagerank), 64)
				if err != nil {
					return err
				}
				pageranks[articleName] = append(pageranks[articleName], Pagerank(float64Pr))
				pagerankIndex[float64Pr] = article
			}
			sort.Sort(sort.Reverse(SortedPageranks(pageranks[articleName])))
			for pagerank, article := range pagerankIndex {
				indexFile.WriteString("** " + article + " - " + strconv.FormatFloat(pagerank, 'f', 6, 64))
			}
		}
	}
	return nil
}

// Words: textcat/n-gram, bayesian+tfidf, pos-tagger(advanced-logic's freeling?)
// TODO: make a basic chatbot which uses Eliza's principle of chatting, the basic outrule for chatting uses Bloom's Taxonomical levels to present data/chat with a user.
// The packages to be used here are word2sentence etc., elasticsearch, and it should utilize the information written to a file in emacs-org format/html-format.
// From here the chatbot is to be written, just to make it J.A.R.V.I.S-like. the chatbot will use ELIZA like chat functionality, e.g explain how one article relates to another by using Bloom's Taxonomical levels (here word2sentence etc. and Elasticsearch will be used).
// The chatbot will load the summaries of the written (from WriteTXT) articles that are in the path of article A, B, C... and their top 5 or 10 pageranking articles.
// The chatbot may start with "Hi, which topics would you like me to chat with you about?". And the chatbot may want to only use a max of 4-7 new words in each message to not strain the short-term memory. It's functionality will be different than the product of WriteTXT function; it will connect the paths between the topics, but the pagerank functionality will work somehow indifferable; IMPORTANT: it will sum up pageranks if a neighboring article has 2 base articles as neighbors. (should maybe also be used for the WriteTXT function).
// Create a neural network for the chatbot with a light prediction analysis algorithm (prediction algorithm trained by user's behaviour) to make it smarter.. Maybe the neural network is not needed since Dijkstra and Pagerank makes up for it; the shortest part and the top pageranking path. Maybe add a neural network to find the path from shortest path and top pageranking path to make it better to chat with humans rather than machines(?).
// Here I may also add a xml dump file of articles from HowStuffWorks if it provides such to add a functinality to the chatbot to answer "how" questions.
// TODO: if the chatbot functionality works well; hook it up to 1 or more irc channels.
// TODO: make a GUI for Android for the chatbot functionality
