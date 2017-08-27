package GhostWriter

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"github.com/mvryan/fasttag"
	"archive/zip"
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

type Buffer bytes.Buffer
type Section struct {
	Content    Buffer
	References Buffer
	// RakeCands  map[string]Buffer
}
// type LeafPR map[string]Buffer
type PageItems struct {
	Sections map[string]Section

	// NodeID Buffer

	// Weight Buffer

	// Links Buffer

	// Pageranks map[uint8]LeafPRmI

	// Index Buffer

	// Synchronizer sync.Mutex
}

/* type Compressor struct {
	Mutex sync.Mutex
	cBuffer Buffer
	GzipWriter gzip.Writer
	GzipReader gzip.Reader
} */

/* type Index struct {
	Map map[string][]byte
} */
type MarkovChain struct {
	Tokens     [3]string
	TokenRatio int32
}
type MarkovChains []MarkovChain
func (mcs MarkovChains) Sort() {
	for num, _ := range mcs {
		sort.Slice(mcs[num], func(j int, i int) bool {
			if mcs[j].TokenRatio == mcs[i].TokenRatio {
				continue
			}
			return mcs[j].TokenRatio < mcs[i].TokenRatio
		})
	}
}

func GetFilesFromArticlesDir(wikiDirectory string) (files []string, err error) {
	osFileInfo, err := ioutil.ReadDir(wikiDirectory)
	if err != nil { return nil, err }
	for number, _ := range osFileInfo {
		if !osFileInfo[number].IsDir() {
			files = append(files, fileInfo.Name())
		}
	}
	return
}

// This part indexes the wikipedia articles, but by modifying it a little the return #1 object from a map with an array byte to a map with a map with an array byte it can contain the index and the article itself.
func IndexWiki(content *string) (map[string][][]byte), err error) {
	pageRE, err := regexp.Compile("<page>(.+)</page>")
	if err != nil { return nil, err }
	titleRE, err := regexp.Compile("<title>(.+)</title>")
	if err != nil { return nil, err }
	referenceRE, err := regexp.Compile("<ref>(.+)</ref>")
	if err != nil { return nil, err }
	linkRE, err := regexp.Compile("[[(.+)]]")
	if err != nil { return nil, err }
	citeRE, err := regexp.Compile("{{cite(.+)}}")
	if err != nil { return nil, err }

	pages := pageRE.FindAllStringIndex(*content, -1)
	titles := titleRE.FindAllStringIndex(*content, -1)
	references := referenceRE.FindAllStringIndex(*content, -1)
	links := linkRE.FindAllStringIndex(*content, -1)
	cites := citeRE.FindAllStringIndex(*content, -1)

	index = make(map[string][]byte)
	for cnt, _ := range pages {
		index[*content[pages[cnt][0]+6:pages[cnt][1]-7]] = []byte(strconv.Itoa(pages[cnt][0]) + "-" + strconv.Itoa(pages[cnt][1]))
	}
	for cnt, _ := range titles {
		index[*content[titles[cnt][0]+7:titles[cnt][1]-8]] = []byte(strconv.Itoa(titles[cnt][0]) + "-" + strconv.Itoa(titles[cnt][1]))
	}
	for cnt, _ := range references {
		index[*content[references[cnt][0]+5:references[cnt][1]-6]] = []byte(strconv.Itoa(references[cnt][0]) + "-" + strconv.Itoa(references[cnt][1]))
	}
	for cnt, _ := range links {
		if strings.Contains(links[cnt], "|") {
			separatorIndex := strings.Index(links[cnt], "|")
			index[*content[links[cnt][0]+2:links[cnt][1]-3]][2:separatorIndex-1] = append(index[*content[links[cnt][0]+2:links[cnt][1]-3]][:separatorIndex-1], []byte(strconv.Itoa(links[cnt][0]) + "-" + strconv.Itoa(links[cnt][1])))
			continue
		}
		index[*content[links[cnt][0]+2:links[cnt][1]-3]][2:len(links[cnt])-2] = append(index[*file][*content[links[cnt][0]+2:links[cnt][1]-3]], []byte(strconv.Itoa(links[cnt][0]) + "-" + strconv.Itoa(links[cnt][1])))
	}
	for cnt, _ := range cites {
		index[*content[cites[cnt][0]+6:cites[cnt][1]-3]] = []byte(strconv.Itoa(cites[cnt][0]) + "-" + strconv.Itoa(cites[cnt][1]))
	}
	return index, nil
}
func WriteIndexZip(writeDirectory *string, indices map[string][]byte) (err error) {
	indexFile, err := os.Create(*writeDirectory + "/index.zip")
	if err != nil { return err }
	zipWriter, err := zip.NewWriter(indexFile)
	if err != nil { return err }
	for article, index := range indices {
		file, err := zipWriter.Create(article)
		if err != nil { return err }
		_, err := file.Write(index)
	}
	zipWriter.Flush()
	if err = zipWriter.Close(); err != nil { return err }
	return nil
}
func WriteIndex(writeDirectory *string, indices map[string][]byte) (err error) {
	indexFile, err := os.Create(*writeDirectory + "/index.txt")
	if err != nil { return err }
	for elements, _ := range indices {
		for 
	}
	zipWriter.Flush()
	if err = zipWriter.Close(); err != nil { return err }
	return nil
}
func ReadWikiIndices(readDirectory *string) (index map[string][]byte, err error) {
	zipReader, err := zip.OpenReader(*readDirectory + "/index.zip")
	article, index := strings.Split(line, "-")
	indices[indexFile.Name()[:6]][article] = []byte(index)
	if err = osFile.Close(); err != nil { return nil, err }
	return index, nil
}
func ReadWikiXML(readDirectory *string, index map[string][]byte) (articles map[string]PageItems, err error) {
	// var xmlDecoder *xml.Decoder
	files, err := GetFilesFromArticlesDir(readDirectory)
	if err != nil { return nil, nil, err }

	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(buffer)
	var bufferLinks string
	articles = make(map[string]*PageItems)

	reTitle, err := regexp.Compile("[=]{2,5}.{1,50}[=]{2,5}")
	if err != nil { return nil, nil, err }
	reReferences, err := regexp.Compile("<ref>(.+)</ref>")
	if err != nil { return nil, nil, err }
	var page MWPage

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
					page.Revisions[0].Text = strings.Replace(page.Revisions[0].Text, "&gt", ">", -1)
					page.Revisions[0].Text = strings.Replace(page.Revisions[0].Text, "&quot", "\"", -1)
					page.Revisions[0].Text = strings.Replace(page.Revisions[0].Text, ";", "", -1)

					refIndex := reReferences.FindAllStringIndex(page.Revisions[0].Text, -1)
					titleIndex := reTitle.FindAllStringIndex(page.Revisions[0].Text, -1)
					for i := 0; i < len(refIndex); i++ {
						if refIndex[0] != nil {
							if refIndex[0][0] > titleIndex[counter][0] {
								counter++
							}
							_, err = gzipWriter.Write([]byte(page.Revisions[0].Text[
								refIndex[i][0]:
								refIndex[1+1][0]
							]))
							if err != nil { return nil, nil, err }
							gzipWriter.Flush()
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
								_, err = gzipWriter.Write(page.Revisions[0].
									Text[:titleIndex[0][1]-1])
								if err != nil { return nil, nil, err }
								gzipWriter.Flush()
								articles[page.Title].Sections["Summary"].Content = buffer.Bytes()
								buffer.Reset()
							} else if i < len(titleIndex)-1 {
								_, err = gzipWriter.Write(page.Revisions[0].
									Text[
									titleIndex[i][1]:
									titleIndex[i+1][0]])
								if err != nil { return nil, nil, err }
								gzipWriter.Flush()
								articles[page.Title].Sections[
									page.Revisions[0].Text[
										titleIndex[i][0]:
										titleIndex[i+1][0]
									]
								].Content = buffer.Bytes()
								buffer.Reset()
							} else {
								_, err = gzipWriter.Write(page.Revisions[0].Text[
									titleIndex[i][1]:
									len(page.Revisions[0].Text)])
								if err != nil { return nil, nil, err }
								gzipWriter.Flush()
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
	if err = gzipWriter.Close(); err != nil { return nil, nil, err }
	if err = ioReader.Close(); err != nil { return nil, nil, err }
	return articles, links, nil
}

func LoadMarkovChain() (PTBTagMap map[string]MarkovChains, WordMap map[string]MarkovChains, err error) {
	// var buffer bytes.Buffer
	// gzipWriter := gzip.NewWriter(&buffer)
	osFileInfo, err := ioutil.ReadDir("dependency_treebank")
	if err != nil { return nil, nil, err }
	var counter uint8 = 0
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
					// Keep and eye on leak with runtime/pprof on goroutines
					go func() {
						for i := 0; i < 3; i++ {
							if *PTBTagMap[buffer[1][0]] {
								*PTBTagMap[buffer[1][0]].TokenRatio++
								continue
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
			}
		}
		for key, _ := range PTBTagMap {
			go PTBTagMap[key].Sort()
		}
		for key, _ := range WordMap {
			go WordMap[key].Sort()
		}
	}
	return PTBTagMap, WordMap, nil
}

// TODO: brill pos tag the MarkovChains, parse the
// example: github.com/korobool/nlp4go uses this progressive model: use a model (in my case oxford's penn tree bank and brill pos tags), train the averaged perceptron (I will use the github.com/sjwhitworth/golearn/base to parse my data into csv format), and then use the averaged perceptron to tag a text. - But for now, use github.com/kamildrazkiewicz/go-stanford-nlp
/* func TagArticle(InputPTBTagMap map[string]*MarkovChains, articles map[string]PageItems) (PTBTagMap map[string]*MarkovChains, WordMap map[string]*MarkovChains, err error) {
	for key, item := range articles {
		for _, text := range articles[key].Sections {
			words := fasttag.WordsToSlice(text)
			for _, word := range words {
				
			}
		}
	}
} */
func (articles map[string]PageItems) Rake() (err error) {
	var readBuffer bytes.Buffer
	gzipReader := gzip.NewReader(&readBuffer)
	var writeBuffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&writeBuffer)
	var words []string

	for key, item := range articles {
		for _, section := range item.Sections {
			gzipReader.Read(item.Sections[sectionTitle])
			gzipReader.Flush()
			rakeCands := rake.RunRake(string(buffer.Bytes()))
			for key, val := range rakeCands {
				gzipWriter.Write(strconv.FormatFloat(val, 'f', -1, 64))
				gzipWriter.Flush()
				section.RakeCands[key] = writeBuffer.Bytes()
				writeBuffer.Reset()
			}
			gzipReader.Reset()
		}
	}
	if err = gzipReader.Close(); err != nil { return err }
	if err = gzipWriter.Close(); err != nil { return err }
}

// TODO
func (pi PageItems) MakeMarkovChainModels(input string) (Trigrams map[string][][]byte err error) {
	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	POSModel := make(map[string][][3]string)
	
}

// TODO
func (articles map[string]PageItems) TfIdf(input *string) (RelevantDocs map[string]string, err error) {
	RelevantDocs := make(map[string]string)
	nouns := make(map[string]string)
	words := fasttag.WordsToSlice(input)
	postags := fasttag.BrillTagger(words)
	var good []string
	for number, _ := range words {
		if postags[number][:1] == "N" {
			good = append(good, words[number])
		}
	}
	for key, _ := range articles {
		for _, section := range articles[key.Sections {

		}
	}
}

// TODO: Implement the 2 underlying methods into the gzip read/write calls.
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
		var indexBuildBuffer buffer bytes.Buffer
		gzipWriter := gzip.NewWriter(&buffer)
		for _, file := range files {
			file, err := os.Open("/run/media/naamik/Data/articles/" + file)
			ArrByteContent, err := ioutil.ReadAll(file)
			if err != nil { panic(err) }
			index, err := IndexWiki(string(ArrByteContent))
			if err != nil { panic(err) }
			err = WriteIndex("/run/media/naamik/Data/articles/index", file + ".index", index)
			if err != nil { panic(err) }
		}
	}
	index, err := ReadWikiIndices("/run/media/naamik/Data/articles/index")
	if err != nil { panic(err) }
	for _, article := range *articles {

	}
}

func (section Section) Decompress(wg sync.WaitGroup, buffer bytes.Buffer, gzipWriter gzip.Reader) (data []byte, err error) {
	wg.Wait()
	var buffer bytes.Buffer
	gzipReader := gzip.NewReader(&buffer)
	_; err = gzipReader.Read(section.Content)
	if err != nil { return nil, err }
	gzipReader.Flush()
	gzipReader.Reset()
	// gzipReader.Close()
	return buffer.Bytes(), nil
}
func (section Section) Compress(data []byte, wg sync.WaitGroup, gzipWriter gzip.Writer) (err error) {
	wg.Wait()
	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	_, err = gzipWriter.Write(data)
	if err != nil { return err }
	gzipWriter.Flush()
	gzipReader.Reset()
	// err = gzipWriter.Close()
	if err != nil { return err }
	section.Content = buffer.Bytes()
	return nil
}
