package myTools
 /* 
		Copyright (C) 2015-2017 Vittus Mikiassen
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
import (
	"io"
	"io/ioutil"
	"os"
	"encoding/gob"
	"strconv"
	"fmt"
	"bytes"
	"encoding/xml"
	"strings"
	"regexp"
)
type Section struct {
	Content []byte
	References []byte
}
type LeafPR map[string][]byte
type PageItems struct {
	// Sections, title indicated by key, item is content/text
	Sections map[string]Section

	// The NodeID, used for graphing/treeing
	NodeID []byte

	// Weight/distance from root
	Weight []byte

	// Links from this article, used to collect them for the MySQL Db, after that the program will use them to utilize the Db for Dijkstra's algorithm and the Pagerank algorithm.
	Links []byte

	Pageranks map[uint8]LeafPR

	Index []byte
}
type Tag string
type Word string
// type Occurence uint
// type Probability float64
type MarkovChainEntities struct {
	Word string
	Tag WordTag
	tmpOccurence uint // This is stored as constOccurence when the amount is determined.
}
type HMMEntities struct {
	Tags [3]string
	tmpTagByTagOccurence uint
}
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
func ReadWikiXML(readDirectory string, filesIndices map[string][]byte) (articles map[string]PageItems, links []byte, err error) {
	files, err := GetFilesFromArticlesDir(readDirectory)
	if err != nil { return nil, nil, err }
	// sections = make(map[string][]byte)
	page := wikiparse.Page{}
	
	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)
	var bufferLinks string

	for _, xmlFile:= range files {
		ioReader, err := os.Open(readDirectory + "/" + xmlFile)
		decoder := xml.NewDecoder(ioReader)
		if err != nil { return nil, nil, err }
		
		for _, offset := range filesIndices {
			parts := strings.Split(string(offset), "-")
			start, err := strconv.ParseInt(parts[0], 10, 64)
			if err != nil { return nil, nil, err }
			_, err = ioReader.Seek(start, io.SeekStart)
			if err != nil { return nil, nil, err }

			err = decoder.Decode(&page)
			if err != nil { return nil, nil, err }

			for i := 0; i < len(page.Revisions); i++ {
				if page.Revisions[i].Text != "" {
					articles[page.Title].Content, err = GetSections(page.Revisions[i].Text, page.Title)
					if err != nil { return nil, nil, err }
					tempLinks := wikiparse.FindLinks(page.Revisions[i].Text)
					for num, link := range tempLinks {
						bufferLinks += link
						if num < len(tempLinks) { bufferLinks += "-" }
					}
					err = encoder.Encode(bufferLinks)
					if err != nil { return nil, nil, err }
					articles[page.Title]["links"] = buffer.Bytes()
					buffer.Reset()
				}
			}
		}
	}
	return articles, links, nil
}
func GetSections(page, title string) (Content PageItems.Content, err error) {
	// TODO: Depending on the hardware limits use 1 loop for both regex searches.
	var lastRefIndex uint // TODO: If a lastRefIndex overflow error should occur increase the buffer length or algorithm to be sequential.
	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)
	reTitle, err := regexp.Compile("[=]{2,5}.{1,50}[=]{2,5}")
	if err != nil { return }
	reReferences, err := regexp.Compile("<ref>\\w+</ref>")
	if err != nil { return }
	for i := 0; i < len(refIndex); i++ {
		// TODO: Remove the references from the Content.
		if refIndex[i] != nil {
			if i > 0 {
				if refIndex[i][1] - 1 == refIndex[i - 1][0] - 1 {
					encoder.Encode([]byte(page[:refIndex[i][1]-1]))
					Content.References[[]byte(strconv.Itoa(refIndex[i][0] - 1))] = buffer.Bytes()
				} else {
					lastRefIndex = refIndex[i][0]
					encoder.Encode([]byte(page[:refIndex[i][1]-1]))
					Content.References[[]byte(strconv.Itoa(refIndex[i][0] - 1))] = buffer.Bytes()
				}
			} else {
				// TODO: Might want to exchange i with 0.
				encoder.Encode([]byte(page[:refIndex[i][1]-1]))
				Content.References[[]byte(strconv.Itoa(refIndex[i][0] - 1))] = buffer.Bytes()
			}
			page[refIndex[i][0] - 1:refIndex[i][1] - 1] = ""
		}
		buffer.Reset()
	}
	titleIndex := reTitle.FindAllStringIndex(page, -1)
	for i := 0; i < len(titleIndex)-1; i++ {
		refIndex := reReferences.FindAllStringIndex(titleIndex[i], -1)
		if titleIndex[i] != nil {
			if i == 0 {
				encoder.Encode([]byte(page[:titleIndex[i][1]-1]))
				Content.Sections["Summary"].Content = buffer.Bytes()
			} else if i < len(titleIndex) - 1 {
				encoder.Encode([]byte(page[titleIndex[i][1]:titleIndex[i+1][0]]))
				Content.Sections[page[titleIndex[i][0]:titleIndex[i][1]]].Content = buffer.Bytes()
			} else {
				encoder.Encode([]byte(page[titleIndex[i][1]:len(page)]))
				Content.Sections[page[titleIndex[i][0]:titleIndex[i][1]]].Content = buffer.Bytes()
			}
			buffer.Reset() // TODO: Depending on the CPU only reset this once and use another buffer for the gob encoder.
		}
	}
	return
}
// TODO: convert the uint item from the chains map to []byte when it is determined how many occurrences of that specific chain there are.
func InitHMMChain() (chains map[MarkovEntities](map[MarkovChainEntities](map[MarkovChainEntities][]byte)), sortedHMMChains map[HMMEntities][]byte, err error) {
	chains = make(map[Word](map[Tag]MarkovChain))
	var buffer [3]MarkovEntities
	var counter uint8 = 0
	var max uint = 0
	var files []string
	sortedChain := make(map[Tag][]string)
	var tagBuffer HMMEntities
	osFileInfo, err := ioutil.ReadDir("dependency_treebank")
	if err != nil { return nil, err }
	for _, file := range osFileInfo {
		if !file.IsDir() {
			files = append(files, file.Name())
			osFile, err := os.Open("dependency_treebank/" + file)
			if err != nil { return nil, err }
			defer osFile.Close()
			bufioReader := bufio.NewReader(osFile)
			for {
				line, _, err  := bufioReader.ReadLine()
				if err == io.EOF {
					break
				} else if err != nil {
					return nil, err
				}
				symbols := strings.Split(string(line), "	")
				max++
				if len(symbols) == 1 { continue }
				counter++
				buffer[counter].Word, buffer[counter].Tag = symbols[0], symbols[1]
				if counter == 3 {
					if !chains[buffer[0]][buffer[1]][buffer[2]] {
						chains[buffer[0]][buffer[1]][buffer[2]] = 1
					} else {
						chains[buffer[0]][buffer[1]][buffer[2]]++ 
					}
					counter = 0
				}
				buffer
			}
		}
	}
	for _, word := range chains {
		for _, tag := range word {
			counter++
			chains[word][tag].Probability = chains[word][tag].Occurence / max
			tagBuffer.Tags[counter] = tag
			if counter == 3 {
				if sortedHMMChains[
			}
		}
	}
	return chains, nil
}
func TagArticle(articles []string) (err error) {
	indices, err := ReadWikiIndices("articles", articles)
	if err != nil { return err }
	articles, _, err := ReadWikiXML("articles", indices)
	if err != nil { return err }
	HMMChain, err := InitHMMChain()
	
}
/* 
TODO: when the final text has to be written, replace the xml entities with the actual textual entities
page = strings.Replace(page, "&lt", "<")
page = strings.Replace(page, "&gt", "<")
page = strings.Replace(page, "&quot", "\"")
page = strings.Replace(page, ";", "")
*/
