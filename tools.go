package GhostWriter

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
	"github.com/mvryan/fasttag"
	"github.com/Obaied/rake"
	// "github.com/gen2brain/go-unarr"
	// "github.com/alixaxel/pagerank"
	// "github.com/jbrukh/bayesian"
	// "github.com/neurosnap/sentences"
	// _ "github.com/go-sql-driver/mysql"
	// "database/sql"
	// "golang.org/x/text/search"
	// "legacy/rosettacode/dijkstra"
)
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
type SEPostHistories struct {
	ChiRow []*PostHistory `xml:" row,omitempty" json:"row,omitempty"`
}

type SEPostLink struct {
	AttrCreationDate string `xml:" CreationDate,attr"  json:",omitempty"`
	AttrId string `xml:" Id,attr"  json:",omitempty"`
	AttrLinkTypeId string `xml:" LinkTypeId,attr"  json:",omitempty"`
	AttrPostId string `xml:" PostId,attr"  json:",omitempty"`
	AttrRelatedPostId string `xml:" RelatedPostId,attr"  json:",omitempty"`
}
type SEPostLinks struct {
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

type Chain struct {
	IndexNearestReferences [2][2]bytes.Buffer
	NGram []bytes.Buffer
}
type Section struct {
	Content bytes.Buffer
	References []bytes.Buffer
	Links []bytes.Buffer
	TopNouns [10][2]bytes.Buffer
	XKeyChain map[string]Chain
}
type Page struct {
	// Content    bytes.Buffer
	Offset bytes.Buffer
	ID bytes.Buffer
	Sections map[string]Section
}
type Articles struct {
	Title bytes.Buffer
	Offset bytes.Buffer
}

// Read all names of Bzipped Wikimedia XML files from "articles" dir.
func GetFilesFromArticlesDir(directory *string) (files []string, err error) {
	osFileInfo, err := ioutil.ReadDir("/run/media/naamik/Data/articles")
	if err != nil { return nil, err }
	for _, fileInfo := range osFileInfo {
		if !fileInfo.IsDir() {
			files = append(files, fileInfo.Name())
		}
	}
	return
}

func DecompressBZip(directory, file *string) (ioReader io.Reader, fileSize int64, err error) {
	osFile, err := os.Open(directory + "/" + file)
	if err != nil { return nil, -1, err }
	fileStat, err := osFile.Stat()
	if err != nil { return nil, fileStat.Size(), err }
	ioReader = bzip2.NewReader(osFile)
	return ioReader, fileStat.Size(), nil
}

// Reads Wikipedia articles from a Wikimedia XML dump bzip file, return the Article with titles as map keys and PageItems (Links, Sections and Text) as items - Also add Section "See Also"
func ReadWikiXML(directory, file *string) (pages map[string]Page, err error) {
	// linkRE := regexp.MustCompile(`\[\[([^\|\]]+)`)
	// refRE := regexp.MustCompile("<ref>(.+)</ref>")

	var page MWPage
	files, err := GetFilesFromArticlesDir()
	if err != nil { return nil, err }

	var page MWPage

	var section string
	var counter uint8 = 0

	for file, _ := range filesIndices {
		ioReader, err := DecompressBZip(directory, *file)
		if err != nil { return nil, err }
		xmlDecoder := xml.NewDecoder(ioReader)
		err = xmlDecoder.Decode(&page)
		if err != nil { if strings.EqualFold(err.Error(), io.EOF.Error()) { break } else { return nil, err } }
		if page.Revisions[i].Text != "" {
			page.Revisions[i].Text = strings.Replace(page.Revisions[0].Text, "&lt", "<", -1); page.Revisions[0].Text = strings.Replace(page.Revisions[0].Text, "&gt", ">", -1); page.Revisions[0].Text = strings.Replace(page.Revisions[0].Text, "&quot", "\"", -1); page.Revisions[0].Text = strings.Replace(page.Revisions[0].Text, ";", "", -1)
			for refNum := len(refIndex); refNum > -1; refNum-- {
				page[title].Sections[section].References = append(page[title].Sections[section].References, page[title].Sections[section].Content[refIndex[refNum][0]:refIndex[refNum][1]])
				page[title].Sections[section].Content[refIndex[refNum][0]:refIndex[refNum][1]] = []byte("[[r:" + strconv.Itoa(refNum) + "]]")
			}
			linkIndex := linkRE.FindAllStringIndex(page[title][section].Content, -1)
			for linkNum := len(linkIndex); linkNum > -1; linkNum-- {
				link := strings.Split(page[title][section].Content[linkIndex[linkNum][0]:refIndex[linkNum][1]], "|")[2:][0]
				page[title].Sections[section].Links = append(page[title].Sections[section].Links, []byte(link))
				page[title].Sections[section].Content[linkIndex[linkNum][0]:linkIndex[linkNum][1]] = []byte("[[l:" + strconv.Itoa(linkNum) + "]]")
			}
			pages[page.Title].GetSections(*page.Revisions[i].Text)
			err = pages[page.Title].TagArticle()
			if err != nil { return err }
			pages[page.Title].Rake()
			pages[page.Title].ID = []byte(strconvItoa(int(page.Revisions[i].ID)))
		}
	}
	return
}
// Gets sections from a wikipedia article, page is article content, title is article title
func (page Page) GetSections(content *string) {
	re := regexp.MustCompile("[=]{2,5}.{1,50}[=]{2,5}")
	if err != nil { return }
	index := re.FindAllStringIndex(page, -1)
	for i := 0; i < len(index)-1; i++ {
		if index[i] != nil {
			if i == 0 {
				page.Sections["Summary"].Content = []byte(*content[:index[i][1]-1])
			} else if i < len(index)-1 {
				page.Sections[page[index[i][0]:index[i][1]]].Content = []byte(*content[index[i][1]:index[i+1][0]])
			} else {
				page.Sections[page[index[i][0]:index[i][1]]].Content = []byte(*content[index[i][1]:len(*content)])
			}
		}
	}
	return
}

func BackwardStringIndex(text, substring string) (index int) {
	for ii := 0; ii < 3; ii++ {
		for iii := len(text); iii > 0; iii--  {
			for iiii := 0; iiii < len(substring) - 1; iiii++ {
				if text[len(text) - iii:len(text) - iii - 1] == substring[iiii:iiii + 1] && ii == 2 {
					return len(text) - iii
				}
			}
		}
	}
	return -1
}
func (page Page) TagArticle() (err error) {
	// ngramRE, err := regexp.Compile("(;:,.!?)\\s(\\w\\d)+(;:,.!?)\\s+(\\w\\d)+(;:,.!?)\\s(\\w\\d)+(;:,.!?)\\s")
	// titleRE := regexp.MustCompile("[=]{2,5}.{1,50}[=]{2,5}")
	// linkRE := regexp.MustCompile(`\[\[([^\|\]]+)`)
	// refRE := regexp.MustCompile("<ref>(.+)</ref>")
	var offset int = 0
	var wordIndex int
	var words, tags []string
	var preSentence, sufSentence []byte = []byte(-1), []byte(-1)
	// the sane choice here would be to load up a lexicon based tagger, but for now I let the runtime manage the package implementation type of tagging. - After all, it is word-ending based, much less memory required, and quite cheap cpu wise for a single operation.
	for title, _ := range page {
		for section, _ := range page[title].Sections {
			words = fasttag.WordsToSlice(string(page[title].Sections[section]))
			tags = fasttag.BrillTagger(words)

			for num, word := range words {
				wordIndex = strings.Index(text, word)
				/* page[title].Sections[section].XKeyChain[word] = Chain{
					Token: tags[num],
					TokenFrequency: float64(1) / float64(len(words)),
				} */
				// sentenceIndex = strings.IndexAny(text[:wordIndex+len(word)+sentenceIndex], ";:,.!?")
				// sentenceIndex = BackwardStringIndex(text[:wordIndex + len(word) + sentenceIndex], ";:,.!?")
				preSentence = []byte(BackwardStringIndex(text[:wordIndex], ";:,.!?"))
				// section.XKeyChain[word].Offset[0] = append(WordMap[word].NGram, text[sentenceIndex+1:wordIndex-sentenceIndex])
				for i := 0; i < 3; i++ {
					sufSentence = strings.IndexAny(text[
						wordIndex +
							len(word) +
							strconv.Atoi(sufSentence) + 1:],
						";:,.!?")
				}
				page[title].Sections[section].XKeyChain[word] = append(page[title].Sections[section], text[preSentence:sufSentence + len(word)])
			}
		}
	}
	break
}
func (page Page) Rake() {
	for title, _ := range page {
		for section, _ := range page[title].Sections {
			rakeCands := rake.RunRake(string(page[title].Sections[section].Content))
			for num, pair := range rakeCands {
				index := strings.Index(string(page[title].Sections[section].Content), pair.Key)
				page.Sections[section].TopNouns = append(section.TopNouns,
					[2]bytes.Buffer{
						[]byte(strconv.Itoa(index)),
						[]byte(strconv.Itoa(index + len(pair.Key))),
					},
				)
				if num == 9 { break }
			}
		}
	}
}

func WriteDB(directory *string, articles map[string]Page) (err error) {
	indexFile, err := os.OpenFile(*directory, "/" + "index.txt", os.O_APPEND, 0666)
	if err != nil { return err }
	defer indexFile.Close()
	contentFile, err := os.OpenFile(*directory + "/" + "content.dat", os.O_APPEND, 0666)
	if err != nil { return err }
	defer contentFile.Close()
	fileStats, err := contentFile.Stat()
	var contentFileLength bytes.Buffer = []byte(strconv.Itoa(int(fileStats.Size())))
	var nounNum uint8 = -1
	var buffer bytes.Buffer
	compressor := gzip.NewWriter(&buffer)
	for title, _ := range articles {
		buffer.Write(articles[title].ID + "|")
		// buffer.Write("|")
		for section, _ := range articles[title].Sections {
			buffer.Write(section)
			buffer.Write("::")
			compressor.Write(articles[title].Sections[section].Content)
			for refNum, _ := range articles[title].Sections[section].References {
				compressor.Write(articles[title].Sections[section].References[refNum])
				if refNum < len(articles[title].Sections[section].References) - 1 { buffer.Write(":") }
			}
			buffer.Write("::")
			for linkNum, _ := range articles[title].Sections[section].Links {
				compressor.Write(articles[title].Sections[section].Links[linkNum])
				if linkNum < len(articles[title].Sections[section].Links) - 1 { buffer.Write(":") }
			}
			buffer.Write("::")
			for key, _ := range articles[title].Sections[section].XKeyChain {
				nounNum++
				buffer.Write(key)
				if nounNum < 9 { buffer.Write(":") }
				compressor.Write(Sections[section].XKeyChain[key])
				if nounNum < 9 { buffer.Write("-") }
			}
			buffer.Write("|")
		}
		fmt.Fprintln(contentFile, buffer.Bytes())
		intContentFileLength, err := strconv.Atoi(contentFileLength)
		if err != nil { return err }
		buffer.Reset()
		err = compressor.Reset(&buffer)
		if err != nil { return err }
		contentFileLength = []byte(strconv.Itoa(intContentFileLength + int(len(buffer.Bytes()))))
		fmt.Fprintln(indexFile, title + ":" + contentFileLength)
	}
	return nil
}
func ReadIndex(directory *string) (index map[string](map[string]bool), err error) {
	indexFile, err := os.Open(*directory + "/index.txt")
	bufioReader := bufio.NewReader(indexFile)
	for {
		line, _, err := bufioReader.ReadLine()
		if err != nil {
			if err == io.EOF { break }
			return nil, err
		}
		parts := strings.Split(string(line), ":")
		index[parts[0]][parts[1]] = true
	}
	return index, nil
}
func (page Page) DecompressContent(directory *string, offset *int64) (err error) {
	file, err := os.Open(*directory + "/" + "contentfile.dat")
	if err != nil { return err }
	_, err = file.Seek(*offset, -1)
	if err != nil { return err }
	/* var buffer bytes.Buffer
	decompressor := gzip.NewReader(&buffer)
	bufioReader, err := bufio.NewReader(file) */
	line, _ , err := bufioReader.ReadLine()
	if err != nil { return err }
	sections := strings.Split(string(line), "|")
	for i := 1; i < len(sections); i++ {
		// for _, section := range sections {
		sectionElements := strings.Split(section[i], "::")
		for num, _ := range sectionElements {
			page[sectionElements[0]].Content = []byte(sectionElements[1])
		}
	}
	return nil
}
// The data stays compressed
func (page Page) GetMetaData(directory *string, offset *int64) (err error) {
	file, err := os.Open(*directory + "/" + "contentfile.dat")
	if err != nil { return err }
	_, err = file.Seek(*offset, -1)
	if err != nil { return err }
	/* var buffer bytes.Buffer
	decompressor := gzip.NewReader(&buffer) */
	bufioReader, err := bufio.NewReader(file)
	line, _ , err := bufioReader.ReadLine()
	if err != nil { return err }
	sections := strings.Split(string(line), "|")
	page.ID = []byte(sections[0])
	for i := 1; i < len(sections); i++ {
		sectionElements := strings.Split(sections[i], "::")
		// page[section[0]].Content = []byte(section[1])
		page[sectionElements[0]].References = strings.Split(sectionElements[1], ":")
		page[sectionElements[0]].Links = strings.Split(sectionElements[2], ":")
		xKeyChainParts := strings.Split(sectionElements[3], "-")
		for i := 0; i < len(xKeyChainParts); i++ {
			absXKeyChainParts := strings.Split(xKeyChainParts, ":")
			page[sectionElements[0]].XKeyChain[absXKeyChainParts[0]] = absXKeyChainParts[1]
		}
	}
	return nil
}

func (section Section) TfIdf(input *string) float64 {
	words := fasttag.WordsToSlice(*input)
	postags := fasttag.BrillTagger(words)
	var good []string

	content := string(section.Content)
	for _, nounOffset := range section.TopNouns {
		good = append(good, strconv.Atoi(content[
			strconv.Atoi(nounOffset[0]):
			strconv.Atoi(nounOffset[1])]))
	}
	return float64(len(good)) / float64(len(fasttag.WordsToSlice(*input)))
}

func main() {
	// TODO: calculate how big a graph can be made for pageranking.
	
}
