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
	// "strings"
	"strconv"
	"regexp"
	// "sync"
	"errors"
	"github.com/dustin/go-wikiparse"
	// "github.com/Professorq/dijkstra"
	//"github.com/alixaxel/pagerank"
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	// "github.com/Masterminds/squirrel"
	//"testing"
)
type PageItems struct {
	Sections map[string]string // Sections (Text) from a wiki article
	NodeID uint64
	Pagerank [7]float64
	Links []string
	//reftohere []string
	//pagerank float64
}
func main() {
	// var articles map[string]PageItems
	// articles = make(map[string]PageItems)
	// var graph map[string][]string
	/* TODO:
		Add functions to read wiki pages, use SqlInsert, an then Dijkstra to load shortetst path when all wiki pages
		and dijkstra graph is loaded into the mysql database.
		When nodes are loaded load pagerank of all links within a depth of 7.
		Load the shortest path from the top 3 of links between each node within a depth of 7.
		From the shortest path take the path with highest pagerank of all the nodes' links between one another if within a depth
		of 7, else just the one with highest pagerank. */
	/* db, err := sql.Open(
		"mysql",
		"root:root@tcp(localhost:3311)/sample")
	if err != nil { panic(err) }
	defer db.Close() */
	err := CreateDB()
	if err != nil { panic(err) }
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
		// Add links in article to our articles - TODO: update to make it fit with dijkstra.go
		links := wikiparse.FindLinks(page.Revisions[i].Text)
		for _, link := range(links) {
			tmp.Links = append(tmp.Links, link)
		}
		articles[page.Title] = tmp
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

func SqlInsert(db *sql.DB, articles map[string]PageItems) (err error) {
	for title, items := range articles {
		// sections := make([]string, len(items.Sections))
		_, err = db.Exec("CREATE TABLE " + title + " (NodeID int, Pagerank_1 float, Pagerank_2 float, Pagerank_3 float, Pagerank_4 float, Pagerank_5 float, Pagerank_6 float, Pagerank_7 float")
		_, err = db.Exec("INSERT INTO " + title + "(NodeID, Pagerank_1, Pagerank_2, Pagerank_3, Pagerank_4, Pagerank_5, Pagerank_6, Pagerank_7, ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", items.NodeID, items.Pagerank[0], items.Pagerank[1], items.Pagerank[2], items.Pagerank[3], items.Pagerank[4], items.Pagerank[5], items.Pagerank[6])
		if err != nil { return err }
		for sectionTitle, sectionBody := range items.Sections {
			_, err = db.Exec("ALTER TABLE " + title + " ADD " + sectionTitle + " text")
			if err != nil { return err }
			_, err = db.Exec("INSERT INTO " + title + "(" + sectionTitle + ") VALUES (?)", sectionBody)
			if err != nil { return err }
		}
	}
	return nil
}
func SqlSelect(db sql.DB, table []string) (articles map[string]PageItems, err error) {
	var tmp PageItems
	for _, title := range table {
		rows, err := db.Query("SELECT * FROM " + title);
		if err != nil { return nil, err }
		columns, err := rows.Columns()
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for rows.Next() {
			for i, _ := range columns {
				valuePtrs[i] = &values[i]
			}
			rows.Scan(valuePtrs...)
			
			for i, _ := range columns {
				tmp = articles[title]
				switch columns[i] {
				case "NodeID":
					if values[i] != nil {
						tmp.NodeID, err = strconv.ParseUint(values[i].(string), 10, 32)
						if err != nil { return nil, err }
					}
				case "Pagerank_1":
					if values[i] != nil {
						fmt.Println("case \"Pagerank\": columns[i]:\t", columns[i], values[i])
						tmp.Pagerank[0], err = strconv.ParseFloat(values[i].(string), 32)
						if err != nil { return nil, err }
					}
				case "Pagerank_2":
					if values[i] != nil {
						fmt.Println("case \"Pagerank\": columns[i]:\t", columns[i], values[i])
						tmp.Pagerank[1], err = strconv.ParseFloat(values[i].(string), 32)
						if err != nil { return nil, err }
					}
				case "Pagerank_3":
					if values[i] != nil {
						fmt.Println("case \"Pagerank\": columns[i]:\t", columns[i], values[i])
						tmp.Pagerank[2], err = strconv.ParseFloat(values[i].(string), 32)
						if err != nil { return nil, err }
					}
				case "Pagerank_4":
					if values[i] != nil {
						fmt.Println("case \"Pagerank\": columns[i]:\t", columns[i], values[i])
						tmp.Pagerank[3], err = strconv.ParseFloat(values[i].(string), 32)
						if err != nil { return nil, err }
					}
				case "Pagerank_5":
					if values[i] != nil {
						fmt.Println("case \"Pagerank\": columns[i]:\t", columns[i], values[i])
						tmp.Pagerank[4], err = strconv.ParseFloat(values[i].(string), 32)
						if err != nil { return nil, err }
					}
				case "Pagerank_6":
					if values[i] != nil {
						fmt.Println("case \"Pagerank\": columns[i]:\t", columns[i], values[i])
						tmp.Pagerank[5], err = strconv.ParseFloat(values[i].(string), 32)
						if err != nil { return nil, err }
					}
				case "Pagerank_7":
					if values[i] != nil {
						fmt.Println("case \"Pagerank\": columns[i]:\t", columns[i], values[i])
						tmp.Pagerank[6], err = strconv.ParseFloat(values[i].(string), 32)
						if err != nil { return nil, err }
					}
				default:
					if values[i] != nil {
						fmt.Println("default: columns[i]:\t", columns[i], values[i])
						tmp.Sections[columns[i]] = values[i].(string)
					}
				}
				articles[title] = tmp
			}
		}
	}
	return articles, nil
}

// TODO: Comment SqlUpdate and SqlDelete out or update them.
/* func SqlUpdate(table, column string, value map[string]interface{}) (query string, err error) {
	query, _, err = squirrel.Update(table).
    Set(column, value).
		ToSql()
	if err != nil { return "", err }
	return query, nil
}
func SqlDelete(table, column, value string) (query string, err error) {
	query, _, err = squirrel.Delete(value).
		From(table).
		ToSql()
	if err != nil { return "", err }
	return query, nil
} */

// Create database with index of the name of first and last article of each wikipedia file.
// DONE
func CreateDB() (err error) {
	// var articleSections []string
	db, err := sql.Open("mysql", "naamik:glvimia7@tcp(localhost:3306)/wikidb")
	if err != nil { return err }
	files, err := ioutil.ReadDir("D:/")
	if err != nil { return err }
	// var nodeID int
	var titles []string
	for _, file := range files {
		absFile, err := os.Open(file.Name())
		if err != nil { return err }
		/* if strings.EqualFold(file.Name(), "enwiki-latest-abstract.xml") {
		}
		if strings.EqualFold(file.Name(), "enwiki-latest-pagelinks.sql.gz") {

		} */
		// wikijsonin, err := DecompressBZip(items.Name())
		// if err != nil { return err }
		if file.Name() == "enwiki-latest-abstract.xml" {
			articles := make(map[string]PageItems)
			// parser, err := wikiparse.NewParser(wikijsonin)
			// if err != nil { return err }
			parser, err := wikiparse.NewParser(absFile)
			if err != nil { return err }
			for err == nil {
				page, err := parser.Next()
				if err != nil { return err }
				// Make ReadWikiXML add an unique ID to each page (Node ID for Dijkstra's algorithm)..
				titles = append(titles, page.Title)
				articles, err = ReadWikiXML(*page)
				if err != nil { return err }
				if err = SqlInsert(db, articles); err != nil { return err }
			}
			/* for title, _ := range(articles) {
			for sectionName, sectionBody := range(articles[title].Sections) {
				// TODO: make an interface for SqlInsert to automatically detect data types to insert.
				query, err := SqlInsert(title, sectionName, sectionBody)
				if err != nil { return err }
			}
		} */
		} else { continue }
	}
	// if err = writeTitles(titles); err != nil { return err }
	return nil
}

// Read the index of all wikipedia articles into a map with key as article file name and item as first and last article name
/* func ReadDB(articleTitles []string) (articles map[string]PageItems, err error) {
	connection, err := sql.Open("mysql", "naamik:glvimia7@tcp(localhost:3306)/wikidb", eventReceiver)
	if err != nil { panic(err) }
	session := connection.NewSession(eventReceiver)
	for _, title := range articleTitles {
		tmp := articles[title]
		tmp, err = sqlSelect(session, title)
		if err != nil { return nil, err }
		articles[title] = tmp
	}
	return articles, nil
} */

// Writing the articles items in emacs-org format (to write the path from article A to B and their top pageranking links in a presentable format)
func WriteTXT(db io.Writer, articles map[string]PageItems) {
	fWriter := bufio.NewWriter(db)
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

/* func Dijkstra() (path map[int]string) {
	
} */

/* func loadNodes() (graph map[int]dijkstra.Vertex, err error) {
	db, err := sql.Open(
		"mysql",
		"root:root@tcp(localhost:3311)/sample")
	if err != nil { return err }
	rows, err := db.Query("SELECT * FROM "
	columns, err := 
	}
} */

/* func writeTitles(titles []string) (err error) {

} */

// Do I need this??
/* func sqlDelete(session *dbr.Session, row, column, value string) (err error) {
	if _, err := session.DeleteFrom(row).
		Where(dbr.Eq(column, value)).
		Exec(); err != nil { return err }
	return nil
} */
