/*
		Copyright (C) 2017 Vittus Mikiassen

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

package GhostWriter
import (
	"fmt"
	"os"
	"io"
	"io/ioutil"
	"bufio"
	"strconv"
	"bytes"
	"strings"
	"regexp"
	"encoding/gob"
	"encoding/xml"
	"sort"
	"github.com/Obaied/rake"
	// "github.com/dustin/go-wikiparse"
)

// Graciously open sources by github.com/dustin/go-wikiparse under an MIT styled license (10TH of July 2017).
type SiteInfo struct {
	SiteName   string `xml:"sitename"`
	Base       string `xml:"base"`
	Generator  string `xml:"generator"`
	Case       string `xml:"case"`
	Namespaces []struct {
		Key   string `xml:"key,attr"`
		Case  string `xml:"case,attr"`
		Value string `xml:",chardata"`
	} `xml:"namespaces>namespace"`
}
// A Contributor is a user who contributed a revision.
type Contributor struct {
	ID       uint64 `xml:"id"`
	Username string `xml:"username"`
}
// A Redirect to another Page.
type Redirect struct {
	Title string `xml:"title,attr"`
}
// A Revision to a page.
type Revision struct {
	ID          uint64      `xml:"id"`
	Timestamp   string      `xml:"timestamp"`
	Contributor Contributor `xml:"contributor"`
	Comment     string      `xml:"comment"`
	Text        string      `xml:"text"`
}
// A Page in the wiki.
type Page struct {
	Title     string     `xml:"title"`
	ID        uint64     `xml:"id"`
	Redir     Redirect   `xml:"redirect"`
	Revisions []Revision `xml:"revision"`
	Ns        uint64     `xml:"ns"`
}

type Section struct {
	Content []byte
	References []byte
	RakeCands rake.PairList
}
type LeafPR map[string][]byte
type PageItems struct {
	// Sections, title indicated by key, item is content/text
	Sections map[string]*Section

	// The NodeID, used for graphing/treeing
	NodeID []byte

	// Weight/distance from root
	Weight []byte

	// Links from this article, used to collect them for the MySQL Db, after that the program will use them to utilize the Db for Dijkstra's algorithm and the Pagerank algorithm.
	Links []byte

	Pageranks map[uint8]*LeafPR

	Index []byte
}

// These cross-reference each other for input/output correlation.
type TokenChains [][3]string
/* type TokenChains struct {
	Chains Tokens
} */
// From: irc, freenode, #go-nuts, user: pestle, snippet: https://play.golang.org/p/F4ACtG_6cP, modified a bit.
func (tc TokenChains) Len() int { return len(tc) }
func (tc TokenChains) Less(i, j int) bool {
	for i, _ := range tc {
		for k := 0; k < 3; k++ {
				// This was not needed in my code.
				if tc[i][k] == tc[j][k] {
					continue
				}
				return tc[i][k] < tc[j][k]
		}
	}
	return false
}
func (tc TokenChains) Swap(i, j int) { tc[i], tc[j] = tc[j], tc[i] }
/* func (tc TokenChains) SortTokens() {
	for num, _ := range tc {
		sort.Sort(tc[num], func(i, j int) bool {
			for k := 0; k < 3; k++ {
				// This was not needed in my code.
				if tc[i][k] == tc[j][k] { continue }
				return tc[i][k] < tc[j][k]
			}
			return false
		})
	}
} */

func GetFilesFromArticlesDir(wikiDirectory string) (files []string, err error) {
	osFileInfo, err := ioutil.ReadDir(wikiDirectory)
	if err != nil { return nil, err }
	for _, fileInfo := range osFileInfo {
		if !fileInfo.IsDir() {
			files = append(files, fileInfo.Name())
		}
	}
	return
}
// This part indexes the wikipedia articles, but by modifying it a little the return #1 object from a map with an array byte to a map with a map with an array byte it can contain the index and the article itself.
func IndexWiki(content string) (index map[string][]byte, err error) {
	indexRE, err := regexp.Compile("<page>")
	if err != nil { return nil, err }
	titleRE, err := regexp.Compile("<title>(\\w+)</title>")
	if err != nil { return nil, err }

	indexIndices := indexRE.FindAllStringIndex(content, -1)
	if err != nil { return nil, err }
	titleIndices := titleRE.FindAllStringIndex(content, -1)
	if err != nil { return nil, err }

	index = make(map[string][]byte)

	for cnt, _ := range titleIndices {
		index[content[titleIndices[cnt][0] + 7:titleIndices[cnt][1] - 8]] = []byte(strconv.Itoa(indexIndices[cnt][0]) + "-" + strconv.Itoa(indexIndices[cnt][1]))
	}
	return 
}
func WriteIndex(writeDirectory, article string, indices map[string][]byte) (err error) {
	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)
	for fileName, index := range indices {
		file, err := os.Create(writeDirectory + "/" + fileName + ".index")
		if err != nil { return err }
		err = encoder.Encode(index)
		fmt.Fprint(file, buffer)
		buffer.Reset()
		file.Close()
	}
	return
}
func ReadWikiIndices(readDirectory string, articles []string) (index map[string](map[string][]byte), err error) {
	buffer := new(bytes.Buffer)
	decoder := gob.NewDecoder(buffer)
	for _, article := range articles {
		files, err := ioutil.ReadDir(readDirectory + "/" + article + ".index")
		if err != nil { return nil, err }
		for _, file := range files {
			index[file.Name()] = make(map[string][]byte)
			osFile, err := os.Open(readDirectory + "/" + article + ".index")
			if err != nil { return nil, err }
			if content, err := ioutil.ReadAll(osFile); err == nil {
				_, err0 := buffer.Read(content)
				if err0 != nil { return nil, err0 }
			} else { return nil, err }
			if err != nil { return nil, err }
			decoder.Decode(index[file.Name()])
		}
	}
	return
}
func ReadWikiXML(readDirectory string, filesIndices map[string][]byte) (articles map[string]*PageItems, links []byte, err error) {
	var xmlDecoder *xml.Decoder
	files, err := GetFilesFromArticlesDir(readDirectory)
	if err != nil { return nil, nil, err }
	// sections = make(map[string][]byte)
	// page := wikiparse.Page{}
	
	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)
	var bufferLinks string
	articles = make(map[string]*PageItems)
	// tmpSection := make(map[string]Section)

	reTitle, err := regexp.Compile("[=]{2,5}.{1,50}[=]{2,5}")
	if err != nil { return }
	reReferences, err := regexp.Compile("<ref>\\w+</ref>")
	if err != nil { return }
	var page Page
	
	var section string
	var counter uint8 = 0

	for _, xmlFile := range files {
		ioReader, err := os.Open(readDirectory + "/" + xmlFile)
		xmlDecoder := xml.NewDecoder(ioReader)
		if err != nil { return nil, nil, err }
		
		for _, offset := range filesIndices {
			parts := strings.Split(string(offset), "-")
			start, err := strconv.ParseInt(parts[0], 10, 64)
			if err != nil { return nil, nil, err }
			_, err = ioReader.Seek(start, io.SeekStart)
			if err != nil { return nil, nil, err }
			for {
				err = xmlDecoder.Decode(&page)
				if err != nil {
					if strings.EqualFold(err.Error(), io.EOF.Error()) {
						break
					} else {
						return nil, nil, err
					}
				}

				if page.Revisions[0].Text != "" {
					// articles[page.Title].Sections, err = GetSections(page.Revisions[i].Text, page.Title)
					//
					page.Revisions[0].Text = strings.Replace(page.Revisions[0].Text, "&lt", "<", -1)
					page.Revisions[0].Text = strings.Replace(page.Revisions[0].Text, "&gt", "<", -1)
					page.Revisions[0].Text = strings.Replace(page.Revisions[0].Text, "&quot", "\"", -1)
					page.Revisions[0].Text = strings.Replace(page.Revisions[0].Text, ";", "", -1)

					// TODO: Depending on the hardware limits/utilization use 1 loop for both regex searches.
					refIndex := reReferences.FindAllStringIndex(page.Revisions[0].Text, -1)
					titleIndex := reTitle.FindAllStringIndex(page.Revisions[0].Text, -1)
					for i := 0; i < len(refIndex); i++ {
						if refIndex[0] != nil {
							if refIndex[0][0] > titleIndex[counter][0] {
								counter++
								section = page.Revisions[0].Text[
									titleIndex[counter][0]:
									titleIndex[counter][1]-1]
							}
							encoder.Encode([]byte(page.Revisions[0].Text[
								refIndex[i][0]:
								refIndex[1+1][0]]))
							// TODO: implement the assignment of the sections struct in this loop and loop below this one.
							// tmpSection[section].References = buffer.Bytes()
							// articles[page.Title].Sections[section] = Section{[]byte{0}, []byte{0}}
							/* if data, ok := articles[page.Title].Sections[section]; ok {
								data.References = buffer.Bytes()
								articles[page.Title].Sections[section] = data
							} */
							articles[page.Title].Sections[section].References = buffer.Bytes()
						}
						buffer.Reset()
					}
					page.Revisions[0].Text = reReferences.ReplaceAllString(page.Revisions[0].Text, "")
					titleIndex = reTitle.FindAllStringIndex(page.Revisions[0].Text, -1)
					for i := 0; i < len(titleIndex)-1; i++ {
						if titleIndex[0] != nil {
							if i == 0 {
								encoder.Encode([]byte(page.Revisions[0].Text[:titleIndex[0][1]-1]))
								/* if data, ok := articles[page.Title].Sections["Summary"]; ok {
									data.Content = buffer.Bytes()
									articles[page.Title].Sections["Summary"].Content = data.Content
								} */
								articles[page.Title].Sections["Summary"].Content = buffer.Bytes()
							} else if i < len(titleIndex) - 1 {
								encoder.Encode([]byte(page.Revisions[0].Text[
									titleIndex[i][1]:
									titleIndex[i+1][0]]))
								articles[page.Title].Sections[page.Revisions[0].Text[
									titleIndex[i][0]:
									titleIndex[i][1]]].Content = buffer.Bytes()
							} else {
								encoder.Encode([]byte(page.Revisions[0].Text[
									titleIndex[i][1]:
									len(page.Revisions[0].Text)]))
								articles[page.Title].Sections[page.Revisions[0].Text[
									titleIndex[i][0]:
									titleIndex[i][1]]].Content = buffer.Bytes()
							}
							buffer.Reset() // TODO: Depending on the CPU only reset this once and use another buffer for the gob encoder.
						}
					}
					//
				}
			}
		}
	}
	return articles, links, nil
}

// This is implemented inside the GetSections function
/* func StripXMLEntities(articles map[string]PageItems) {
	for _, article := range articles {
	page = strings.Replace(page, "&lt", "<")
	page = strings.Replace(page, "&gt", "<")
	page = strings.Replace(page, "&quot", "\"")
	page = strings.Replace(page, ";", "")
} */
// TODO: convert the uint item from the chains map to []byte when it is determined how many occurrences of that specific chain there are.

// Load the markov chains tagged with the penn treebank tags, repository: https://nlp.stanford.edu/links/statnlp.html
func LoadMarkovChain() (PTBTagMap map[string]*TokenChains, WordMap map[string]*TokenChains, err error) {
	markovEntities := make(map[string]string)
	osFileInfo, err := ioutil.ReadDir("dependency_treebank")
	if err != nil { return nil, nil, err }
	var counter uint8 = 0
	var isInMap bool = false
	var buffer [2][3]string
	
	for _, file := range osFileInfo {
		if !file.IsDir() {
			osFile, err := os.Open("dependency_treebank/" + file.Name())
			if err != nil { return nil, nil, err }
			defer osFile.Close()
			bufioReader := bufio.NewReader(osFile)
			for {
				line, _, err  := bufioReader.ReadLine()
				if err == io.EOF {
					break
				} else if err != nil {
					return nil, nil, err
				}
				symbols := strings.Split(string(line), "	")
				buffer[0][counter] = symbols[0]
				buffer[1][counter] = symbols[1]
				counter++
	
				if counter == 2 { counter = 0 }
			}
			for i := 0; i < len(buffer[0]) - 3; i++ {
				// PTBTagMap[buffer[1][0]].Chains = append(PTBTagMap[buffer[1][0]].Chains, [3]string{buffer[i][i], buffer[i][i + 1], buffer[i][i + 2]})
				*PTBTagMap[buffer[1][0]] = append(*PTBTagMap[buffer[1][0]], [3]string{buffer[i][i], buffer[i][i + 1], buffer[i][i + 2]})
				// WordMap[buffer[i][0]].Chains = append(WordMap[buffer[i][0]].Chains, [3]string{buffer[1][i], buffer[1][i + 1], buffer[1][i + 2]})
				*WordMap[buffer[i][0]] = append(*WordMap[buffer[i][0]], [3]string{buffer[1][i], buffer[1][i + 1], buffer[1][i + 2]})
			}
			for key, _ := range PTBTagMap {
				// PTBTagMap[key].SortTokens()
				sort.Sort(PTBTagMap[key])
			}
			for key, _ := range WordMap {
				// WordMap[key].SortTokens()
				sort.Sort(WordMap[key])
			}
		}
	}
	return PTBTagMap, WordMap, nil
}
func NLP(articles map[string]PageItems) (text string, err error) {
	// These 2 below functions should be called from the main function.
	/* indices, err := ReadWikiIndices("articles", articles)
	if err != nil { return nil, err }
	articles, _, err = ReadWikiXML("articles", indices)
	if err != nil { return nil, err } */
	decBuffer := new(bytes.Buffer)
	decoder := gob.NewDecoder(decBuffer)
	encBuffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(encBuffer)
	
	PTBTagMap, WordMap,  err := LoadMarkovChain()
	if err != nil { return "", err }
	for key, item := range articles {
		for sectionTitle, _ := range item.Sections {
			decoder.Decode(articles[key].Sections[sectionTitle].Content)
			articles[key].Sections[sectionTitle].RakeCands = rake.RunRake(string(decBuffer.Bytes()))
			decBuffer.Reset()
		}
	}
	
}
