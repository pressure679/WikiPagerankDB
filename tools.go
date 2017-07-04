package myTools
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
	"github.com/nyxtom/viterbi"
	"github.com/Obaied/rake"
	"github.com/jbrukh/bayesian"
)

type Section struct {
	Content []byte
	References []byte
	RakeCands *[]string
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

// These cross-reference each other for input/output correlation.
type Tokens [][3]string
type TokenChains struct {
	Chains Tokens
}
// From: irc, freenode, #go-nuts, user: pestle, snippet: https://play.golang.org/p/F4ACtG_6cP, modified a bit.
func (tc TokenChains) SortTokens() {
	for num, _ := range tc {
		sort.Slice(tc[num], func(i, j int) bool {
			for k := 0; k < 3; k++ {
				// This was not needed in my code.
				if data[i][k] == data[j][k] {
					continue
				}
				return data[i][k] < data[j][k]
			}
			return false
		})
	}
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
func GetSections(page, title string) (Content PageItems, err error) {
	page = strings.Replace(page, "&lt", "<")
	page = strings.Replace(page, "&gt", "<")
	page = strings.Replace(page, "&quot", "\"")
	page = strings.Replace(page, ";", "")

	// TODO: Depending on the hardware limits/utilization use 1 loop for both regex searches.
	var lastRefIndex uint // TODO: If a lastRefIndex overflow error should occur increase the buffer length or algorithm to be sequential.
	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)
	reTitle, err := regexp.Compile("[=]{2,5}.{1,50}[=]{2,5}")
	if err != nil { return }
	reReferences, err := regexp.Compile("<ref>\\w+</ref>")
	if err != nil { return }
	for i := 0; i < len(refIndex); i++ {
		if refIndex[i] != nil {
			if i > 0 {
				if refIndex[i][1] - 1 == refIndex[i - 1][0] - 1 {
					encoder.Encode([]byte(page[:refIndex[i][1]-1]))
					Content.Sections[[]byte(strconv.Itoa(refIndex[i][0] - 1))].References = buffer.Bytes()
				} else {
					lastRefIndex = refIndex[i][0]
					encoder.Encode([]byte(page[:refIndex[i][1]-1]))
					Content.Sections[[]byte(strconv.Itoa(refIndex[i][0] - 1))].References = buffer.Bytes()
				}
			} else {
				// TODO: Might want to exchange i with 0.
				encoder.Encode([]byte(page[:refIndex[i][1]-1]))
				Content.Sections[[]byte(strconv.Itoa(refIndex[i][0] - 1))].References = buffer.Bytes()
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
// This is implemented inside the GetSections function
/* func StripXMLEntities(articles map[string]PageItems) {
	for _, article := range articles {
	page = strings.Replace(page, "&lt", "<")
	page = strings.Replace(page, "&gt", "<")
	page = strings.Replace(page, "&quot", "\"")
	page = strings.Replace(page, ";", "")
} */
// TODO: convert the uint item from the chains map to []byte when it is determined how many occurrences of that specific chain there are.

func LoadMarkovChain() (PTBTagMap map[string]TokenChains, WordMap map[string]TokenChains, err error) {
	markovEntities := make(map[string]string)
	osFileInfo, err := ioutil.ReadDir("dependency_treebank")
	if err != nil { return nil, err }
	var counter uint8 = 0
	var isInMap bool = false
	var buffer [2][3]string
	
	for _, file := range osFileInfo {
		if !file.IsDir() {
			files = append(files, file.Name())
			osFile, err := os.Open("dependency_treebank/" + file)
			if err != nil { return nil, err }
			defer osFile.Close()
			bufioReader := bufio.NewReader(osFile)
			for {
				counter++
				line, _, err  := bufioReader.ReadLine()
				if err == io.EOF {
					break
				} else if err != nil {
					return nil, nil, err
				}
				symbols := strings.Split(string(line), "	")
				buffer[0] = append(buffer[0], symbols[0])
				buffer[1] = append(buffer[1], symbols[1])
			}
			for i := 0; i < len(buffer[0]) - 3; i++ {
				PTBTagMap[buffer[1]].Chains = append(PTBTagMap[buffer[1]].Chains, buffer[0][i])
				PTBTagMap[buffer[1]].Chains = append(PTBTagMap[buffer[1]].Chains, buffer[0][i + 1])
				PTBTagMap[buffer[1]].Chains = append(PTBTagMap[buffer[1]].Chains, buffer[0][i + 2])
				WordMap[buffer[0]].Chains = append(WordMap[buffer[0]].Chains, buffer[1][i])
				WordMap[buffer[0]].Chains = append(WordMap[buffer[0]].Chains, buffer[1][i + 1])
				WordMap[buffer[0]].Chains = append(WordMap[buffer[0]].Chains, buffer[1][i + 2])
			}
			for key, _ := range PTBTagMap {
				PTBTagMap[key].SortTokens()
			}
			for key, _ := range PTBTagMap {
				WordMap[key].SortTokens()
			}
		}
	}
	return PTBTagMap, WordMap, nil
}
func NLP(articles map[string]PageItems) (text string, err error) {
	indices, err := ReadWikiIndices("articles", articles)
	if err != nil { return nil, err }
	articles, _, err := ReadWikiXML("articles", indices)
	if err != nil { return nil, err }
	PTBTagMap, WordMap,  err := LoadMarkovChain()
	if err != nil { return nil, err }
	// Rake and TF-IDF and produce text between input/output texts based on pageranked input.
	/* const badTags []string = []string{"NN", "PRP", "WP", "EX"}
	const NonWord []string = []string{ }
	var badWords []string
	for absBadTag := range badTags {
		for _, absBadWord := range PTBTagMap[absBadWord].Chains {
			badWords = append(badWords, absBadWord[0])
		}
	} */
	for key, item := range articles {
		for _, sectionTitle := range item.Sections {
			*articles[key].Sections[sectionTitle].RakeCands = rake.RunRake(articles[key].Sections[sectionTitle].Content)
		}
	}
}

