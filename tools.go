package GhostWriter

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"encoding/xml"
	"flag"
	"fmt"
	"github.com/Obaied/rake"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	// "github.com/dustin/go-wikiparse"
	"github.com/mvryan/fasttag"
	"github.com/neurosnap/sentences"
	"github.com/neurosnap/sentences/data"
)

// Graciously open sourced by github.com/dustin/go-wikiparse under an MIT styled license (10TH of July 2017).
type MWSiteInfo struct {
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
type MWContributor struct {
	ID       uint64 `xml:"id"`
	Username string `xml:"username"`
}

// A Redirect to another Page.
type MWRedirect struct {
	Title string `xml:"title,attr"`
}

// A Revision to a page.
type MWRevision struct {
	ID          uint64      `xml:"id"`
	Timestamp   string      `xml:"timestamp"`
	Contributor MWContributor `xml:"contributor"`
	Comment     string      `xml:"comment"`
	Text        string      `xml:"text"`
}

// A Page in the wiki.
type MWPage struct {
	Title     string     `xml:"title"`
	ID        uint64     `xml:"id"`
	Redir     Redirect   `xml:"redirect"`
	Revisions []MWRevision `xml:"revision"`
	Ns        uint64     `xml:"ns"`
}

type SEBadge struct {
	AttrClass string `xml:" Class,attr"  json:",omitempty"`
	AttrDate string `xml:" Date,attr"  json:",omitempty"`
	AttrId string `xml:" Id,attr"  json:",omitempty"`
	AttrName string `xml:" Name,attr"  json:",omitempty"`
	AttrTagBased string `xml:" TagBased,attr"  json:",omitempty"`
	AttrUserId string `xml:" UserId,attr"  json:",omitempty"`
}
type SEBadges struct {
	SEBadgeRow []*SEBadge `xml:" row,omitempty" json:"row,omitempty"`
}

type SEComment struct {
	AttrCreationDate string `xml:" CreationDate,attr"  json:",omitempty"`
	AttrId string `xml:" Id,attr"  json:",omitempty"`
	AttrPostId string `xml:" PostId,attr"  json:",omitempty"`
	AttrScore string `xml:" Score,attr"  json:",omitempty"`
	AttrText string `xml:" Text,attr"  json:",omitempty"`
	AttrUserDisplayName string `xml:" UserDisplayName,attr"  json:",omitempty"`
	AttrUserId string `xml:" UserId,attr"  json:",omitempty"`
}
type SEComments struct {
	ChiRow []*SEComment `xml:" row,omitempty" json:"row,omitempty"`
}

type SEPostHistory struct {
	AttrComment string `xml:" Comment,attr"  json:",omitempty"`
	AttrCreationDate string `xml:" CreationDate,attr"  json:",omitempty"`
	AttrId string `xml:" Id,attr"  json:",omitempty"`
	AttrPostHistoryTypeId string `xml:" PostHistoryTypeId,attr"  json:",omitempty"`
	AttrPostId string `xml:" PostId,attr"  json:",omitempty"`
	AttrRevisionGUID string `xml:" RevisionGUID,attr"  json:",omitempty"`
	AttrText string `xml:" Text,attr"  json:",omitempty"`
	AttrUserDisplayName string `xml:" UserDisplayName,attr"  json:",omitempty"`
	AttrUserId string `xml:" UserId,attr"  json:",omitempty"`
}
type PostHistories struct {
	ChiRow []*PostHistory `xml:" row,omitempty" json:"row,omitempty"`
}

type SEPostLink struct {
	AttrCreationDate string `xml:" CreationDate,attr"  json:",omitempty"`
	AttrId string `xml:" Id,attr"  json:",omitempty"`
	AttrLinkTypeId string `xml:" LinkTypeId,attr"  json:",omitempty"`
	AttrPostId string `xml:" PostId,attr"  json:",omitempty"`
	AttrRelatedPostId string `xml:" RelatedPostId,attr"  json:",omitempty"`
}
type PostLinks struct {
	ChiRow []*SEPostLink `xml:" row,omitempty" json:"row,omitempty"`
}

type SEPost struct {
	AttrAcceptedAnswerId string `xml:" AcceptedAnswerId,attr"  json:",omitempty"`
	AttrAnswerCount string `xml:" AnswerCount,attr"  json:",omitempty"`
	AttrBody string `xml:" Body,attr"  json:",omitempty"`
	AttrClosedDate string `xml:" ClosedDate,attr"  json:",omitempty"`
	AttrCommentCount string `xml:" CommentCount,attr"  json:",omitempty"`
	AttrCreationDate string `xml:" CreationDate,attr"  json:",omitempty"`
	AttrFavoriteCount string `xml:" FavoriteCount,attr"  json:",omitempty"`
	AttrId string `xml:" Id,attr"  json:",omitempty"`
	AttrLastActivityDate string `xml:" LastActivityDate,attr"  json:",omitempty"`
	AttrLastEditDate string `xml:" LastEditDate,attr"  json:",omitempty"`
	AttrLastEditorDisplayName string `xml:" LastEditorDisplayName,attr"  json:",omitempty"`
	AttrLastEditorUserId string `xml:" LastEditorUserId,attr"  json:",omitempty"`
	AttrOwnerDisplayName string `xml:" OwnerDisplayName,attr"  json:",omitempty"`
	AttrOwnerUserId string `xml:" OwnerUserId,attr"  json:",omitempty"`
	AttrParentId string `xml:" ParentId,attr"  json:",omitempty"`
	AttrPostTypeId string `xml:" PostTypeId,attr"  json:",omitempty"`
	AttrScore string `xml:" Score,attr"  json:",omitempty"`
	AttrTags string `xml:" Tags,attr"  json:",omitempty"`
	AttrTitle string `xml:" Title,attr"  json:",omitempty"`
	AttrViewCount string `xml:" ViewCount,attr"  json:",omitempty"`
}
type SEPosts struct {
	ChiRow []*SEPost `xml:" row,omitempty" json:"row,omitempty"`
}

type SETag struct {
	AttrCount string `xml:" Count,attr"  json:",omitempty"`
	AttrExcerptPostId string `xml:" ExcerptPostId,attr"  json:",omitempty"`
	AttrId string `xml:" Id,attr"  json:",omitempty"`
	AttrTagName string `xml:" TagName,attr"  json:",omitempty"`
	AttrWikiPostId string `xml:" WikiPostId,attr"  json:",omitempty"`
}
type SETags struct {
	ChiRow []*SETag `xml:" row,omitempty" json:"row,omitempty"`
}

type SEUser struct {
	AttrAboutMe string `xml:" AboutMe,attr"  json:",omitempty"`
	AttrAccountId string `xml:" AccountId,attr"  json:",omitempty"`
	AttrAge string `xml:" Age,attr"  json:",omitempty"`
	AttrCreationDate string `xml:" CreationDate,attr"  json:",omitempty"`
	AttrDisplayName string `xml:" DisplayName,attr"  json:",omitempty"`
	AttrDownVotes string `xml:" DownVotes,attr"  json:",omitempty"`
	AttrId string `xml:" Id,attr"  json:",omitempty"`
	AttrLastAccessDate string `xml:" LastAccessDate,attr"  json:",omitempty"`
	AttrLocation string `xml:" Location,attr"  json:",omitempty"`
	AttrProfileImageUrl string `xml:" ProfileImageUrl,attr"  json:",omitempty"`
	AttrReputation string `xml:" Reputation,attr"  json:",omitempty"`
	AttrUpVotes string `xml:" UpVotes,attr"  json:",omitempty"`
	AttrViews string `xml:" Views,attr"  json:",omitempty"`
	AttrWebsiteUrl string `xml:" WebsiteUrl,attr"  json:",omitempty"`
}
type SEUsers struct {
	ChiRow []*SEUser `xml:" row,omitempty" json:"row,omitempty"`
}

type Section struct {
	Content    []byte
	References []byte
	RakeCands  map[string][][]byte
}
type LeafPR map[string][]byte
type PageItems struct {
	Sections map[string]*Section

	NodeID []byte

	Weight []byte

	Links []byte

	Pageranks map[uint8]*LeafPR

	Index []byte
}

type MarkovChain struct {
	Tokens     [3]string
	TokenRatio []byte
}
type TokenChains []MarkovChain

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
	var buffer bytes.Buffer
	zw := gzip.NewWriter(&buffer)
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
		_, err = zw.Write([]byte(ndexIndices[cnt][0] + "-" + strconv.Itoa(indexIndices[cnt][1])))
		if err != nil { return nil, err }
		index[content[titleIndices[cnt][0]+7:titleIndices[cnt][1]-8]] = buffer.Bytes()
		if err := zw.Close(); err != nil { return nil, err }
		zw.Reset(&buffer)
	}
	return
}
func WriteIndex(writeDirectory, file string, indices map[string][]byte) (err error) {
	indexFile, err := os.Create("/run/media/naamik/Data/articles/index/" + file + ".index")
	if err != nil { return err }
	for article, index := range indices {
		if err != nil { return err }
		fmt.Fprintln(indexFile, article + "-" + string(index))
	}
	if err = indexFile.Close(); err != nil { return err }
	return nil
}
func ReadWikiIndices(readDirectory string) (indices map[string](map[string][]byte), err error) {
	indexFiles, err := ioutil.ReadDir(readDirectory)
	for _, indexFile := range indexFiles {
		if indexFile.IsDir() { continue }
		if err != nil { return nil, err }
		index[file.Name()] = make(map[string][]byte)
		osFile, err := os.Open(readDirectory + "/" + indexFile.Name())
		if err != nil { return nil, err }
		fileReader := bufio.NewReader(osFile)
		for {
			if line, _, err := fileReader.ReadLine(); err != nil {
				article, index := strings.Split(line, "-")
				indices[indexFile.Name()[:6]][article] = []byte(index)
			} else if err == io.EOF {
				break
			} else { return nil, err }
		}
		if err = osFile.Close(); err != nil { return nil, err }
	}
	return index, nil
}
func ReadWikiXML(readDirectory string, filesIndices map[string](map[string][]byte)) (articles map[string]*PageItems, links []byte, err error) {
	var xmlDecoder *xml.Decoder
	files, err := GetFilesFromArticlesDir(readDirectory)
	if err != nil { return nil, nil, err }

	var buffer bytes.Buffer
	zw := gzip.NewWriter(&buffer)
	var bufferLinks string
	articles = make(map[string]*PageItems)

	reTitle, err := regexp.Compile("[=]{2,5}.{1,50}[=]{2,5}")
	if err != nil { return nil, nil, err }
	reReferences, err := regexp.Compile("<ref>\\w+</ref>")
	if err != nil { return nil, nil, err }
	var page wikiparse.Page

	var section string
	var counter uint8 = 0

	for file, _ := range filesIndices {
		ioReader, err := os.Open(readDirectory + "/" + file)
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
					} else { return nil, nil, err }
				}

				if page.Revisions[0].Text != "" {
					page.Revisions[0].Text = strings.Replace(page.Revisions[0].Text, "&lt", "<", -1)
					page.Revisions[0].Text = strings.Replace(page.Revisions[0].Text, "&gt", "<", -1)
					page.Revisions[0].Text = strings.Replace(page.Revisions[0].Text, "&quot", "\"", -1)
					page.Revisions[0].Text = strings.Replace(page.Revisions[0].Text, ";", "", -1)

					refIndex := reReferences.FindAllStringIndex(page.Revisions[0].Text, -1)
					titleIndex := reTitle.FindAllStringIndex(page.Revisions[0].Text, -1)
					for i := 0; i < len(refIndex); i++ {
						if refIndex[0] != nil {
							if refIndex[0][0] > titleIndex[counter][0] {
								counter++
							}
							_, err = zw.Write([]byte(page.Revisions[0].Text[
								refIndex[i][0]:
								refIndex[1+1][0]
							]))
							if err != nil { return nil, nil, err }
							zw.Flush()
							articles[page.Title].
								Sections[section].
								References = buffer.Bytes()
							buffer.Reset()
						}
					}
					page.Revisions[0].Text = reReferences.ReplaceAllString(page.Revisions[0].Text, "")
					titleIndex = reTitle.FindAllStringIndex(page.Revisions[0].Text, -1)
					for i := 0; i < len(titleIndex)-1; i++ {
						if titleIndex[0] != nil {
							if i == 0 {
								_, err = zw.Write(page.Revisions[0].
									Text[:titleIndex[0][1]-1])
								if err != nil { return nil, nil, err }
								zw.Flush()
								articles[page.Title].Sections["Summary"].Content = buffer.Bytes()
								buffer.Reset()
							} else if i < len(titleIndex)-1 {
								_, err = zw.Write(page.Revisions[0].
									Text[
									titleIndex[i][1]:
									titleIndex[i+1][0]])
								if err != nil { return nil, nil, err }
								zw.Flush()
								articles[page.Title].Sections[
									page.Revisions[0].Text[
										titleIndex[i][0]:
										titleIndex[i+1][0]
									]
								].Content = buffer.Bytes()
								buffer.Reset()
							} else {
								_, err = zw.Write(page.Revisions[0].Text[
									titleIndex[i][1]:
									len(page.Revisions[0].Text)])
								if err != nil { return nil, nil, err }
								zw.Flush()
								articles[page.Title].Sections[
									page.Revisions[0].Text[
										titleIndex[i][0]:
										titleIndex[i][1]
									]
								].Content = buffer.Bytes()
								buffer.Reset()
							}
						}
					}
				}
			}
		}
	}
	if err = zw.Close(); err != nil { return nil, nil, err }
	if err = ioReader.Close(); err != nil { return nil, nil, err }
	return articles, links, nil
}

// This function Loads up 2 markov chain maps containing a sorted 1) Penn Tree Bank, and 2) a trigram of words from sentences. The maps are sorted by a ratio of occurence. The Tree bank has a reference for download here: http://www.nltk.org/nltk_data/ , it is under #56.
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
				line, _, err := bufioReader.ReadLine()
				if err == io.EOF {
					break
				} else if err != nil { return nil, nil, err }
				symbols := strings.Split(string(line), "	")
				buffer[0][counter] = symbols[0]
				buffer[1][counter] = symbols[1]
				counter++

				if counter == 3 {
					counter = 0
					for i := 0; i < 3; i++ {
						if *PTBTagMap[buffer[1][0]] {
							// TODO
							*PTBTagMap[buffer[1][0]].TokenRatio = strconv.ParseUint(*PTBTagMap[buffer[1][0]].TokenRatio, 10, 32)
						}
						*PTBTagMap[buffer[1][0]] = append(*PTBTagMap[buffer[1]],
							&MarkovChain{Tokens: [3]string{
								buffer[0][i],
								buffer[0][i+1],
								buffer[0][i+2]},
							},
						)
						*WordMap[buffer[i][0]] = append(*WordMap[buffer[i]],
							&MarkovChain{Tokens: [3]string{
								buffer[1][i],
								buffer[1][i+1],
								buffer[1][i+2]},
							},
						)
					}
				}
			}
			for key, _ := range PTBTagMap {
				sort.Slice(PTBTagMap[key], func(i int, j int) bool {

				})
			}
			for key, _ := range WordMap {
				sort.Slice(WordMap[key], func(i int, j int) bool {

				})
			}
		}
	}
	return PTBTagMap, WordMap, nil
}

/* func TagArticle(articles map[string]PageItems) (PTBTagMap map[string]*TokenChains, WordMap map[string]*TokenChains, err error) {
	for key, item := range articles {

	}
} */
func (articles map[string]*PageItems) NLP() {
	b, _ := data.Asset("data/english.json")
	training, _ := sentences.LoadTraining(b)
	tokenizer := sentences.NewSentenceTokenizer(training)
	// These 2 below functions should be called from the main function.
	/* indices, err := ReadWikiIndices("articles", articles)
	if err != nil { return nil, err }
	articles, _, err = ReadWikiXML("articles", indices)
	if err != nil { return nil, err } */
	decBuffer := new(bytes.Buffer)
	decoder := gob.NewDecoder(decBuffer)
	encBuffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(encBuffer)
	sents = make(map[string]sentences.RakeCands)
	rakeCands = make(map[string]sentences.RakeCands)
	var absWords []string

	/* PTBTagMap, WordMap, err := LoadMarkovChain

	if err != nil { return "", err } */
	for key, item := range articles {
		for sectionTitle, _ := range item.Sections {
			decoder.Decode(articles[key].Sections[sectionTitle].Content)
			// I am not sure if RakeCands is needed in the struct, but for now keep it.
			sents[sectionTitle].RakeCands = rake.RunRake(string(decBuffer.Bytes()))
			sentences := tokenizer.Tokenize(string(decBuffer.Bytes()))
			decBuffer.Reset()
			for _, sentence := range sentences {
				if strings.ContainsAny(sentence.String(), " ,.-?!\"':;") {
					absWords = strings.Split(strings.Trim(sentence.Text, " ,.-?!\"':;"), " ")
				} else {
					absWords = strings.Split(sentence.Text, " ")
				}
				for i := 0; i < sents[sectionTitle].RakeCands/10; i++ {
					for _, absWord := range absWords {
						// TODO: add a "if strings.Contains(articles[key].Sections[sectionTitle].RakeCands[sectionTitle(iota...)])
						if strings.Contains(articles[key].Sections[sectionTitle].RakeCands[i], absWord) {
							// TODO: The append has to gob encode the appendage var
							articles[key].Sections[sectionTitle].RakeCands[absWord] = append(articles[key].Sections[sectionTitle].RakeCands[absWord], sents[sectionsTitle].RakeCands[i])
							// How to make a text out of these rakecands?
							// - brill fast pos tag on the rakecands and make a perceptron trained by the possible pos nodes, and utilize it on the rakecands - or bayesian sentiment scoring?
						}
					}
				}
			}
		}
	}
}

func main() {
	buildIndex := flag.Bool("-build_index", false, "if the index is not build, provide this as an argument")
	articleFlag := flag.String("-articles", "", "Write the desired articles to use as input for the program, but use _ instead of \" \", and separate the articles with a \",\"")
	flag.Parse()
	if strings.EqualFold(articles, "") {
		fmt.Println("--articles argument must to be non-nil")
		break
	}
	articles := strings.Split(*articleFlag, ",")
	defer recover()
	files, err := GetFilesFromArticlesDir("/run/media/naamik/Data/articles")
	if err != nil { panic(err) }
	if *buildIndex {
		for _, file := range files {
			file, err := os.Open("/run/media/naamik/Data/articles/" + file)
			ArrByteContent, err := ioutil.ReadAll(file)
			if err != nil { panic(err) }
			index, err := IndexWiki(string(ArrByteContent))
			if err != nil { panic(err) }
			err = WriteIndex("/run/media/naamik/Data/articles/index", file+".index", index)
			if err != nil { panic(err) }
		}
	}
	index, err := ReadWikiIndices("/run/media/naamik/Data/articles/index")
	if err != nil { panic(err) }
	for _, article := range *articles {

	}
}
