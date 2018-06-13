package main
import (
	"fmt"
	"os"
	"github.com/mvryan/fasttag"
	"github.com/neurosnap/sentences"
	"github.com/neurosnap/sentences/data"
	"sort"
	"io/ioutil"
	"sync"
)
type Sentence struct {
	Sentence string
	PosTags []string
}
type Noun struct {
	Noun string
	Occurence uint
}
type nounSorter struct {
	nouns []*Noun
	by func(n1, n2 *Noun) bool
}
type ByNoun func(n1, n2 *Noun) bool
func (by ByNoun) Sort(nouns []*Noun) {
	ns := &nounSorter{
		nouns: nouns,
		by: by,
	}
	sort.Sort(ns)
}
func (ns *nounSorter) Len() int {
	return len(ns.nouns)
}
func (ns *nounSorter) Swap(i, j int) {
	ns.nouns[i], ns.nouns[j] = ns.nouns[j], ns.nouns[i]
}
func (ns *nounSorter) Less(i, j int) bool {
	return ns.by(ns.nouns[i], ns.nouns[j])
}
type TriGram struct {
	Nouns [3]*Noun
	Occurence uint
}
type trigramSorter struct {
	trigrams []*TriGram
	by func(n1, n2 *TriGram) bool
}
type ByTriGram func(t1, t2 *TriGram) bool
func (by ByTriGram) Sort(trigrams []*TriGram) {
	ts := &trigramSorter{
		trigrams: trigrams,
		by: by,
	}
	sort.Sort(ts)
}
func (ts *trigramSorter) Len() int {
	return len(ts.trigrams)
}
func (ts *trigramSorter) Swap(i, j int) {
	ts.trigrams[i], ts.trigrams[j] = ts.trigrams[j], ts.trigrams[i]
}
func (ts *trigramSorter) Less(i, j int) bool {
	return ts.by(ts.trigrams[i], ts.trigrams[j])
}
type Pointers struct {
	Nouns []*Noun
	TriGrams []*TriGram
}
type Graph struct {
	Sentences []Sentence
	Nouns []Noun
	TriGrams []TriGram

	qaNouns map[string]uint
	qaNounToSentences map[string][]uint

	qaTriGrams map[[3]string]uint
	qaNounToTriGram map[string]uint

	Final Pointers
	Print []*Sentence
}
func main() {
	var wg sync.WaitGroup
	var mygraph Graph
	var trigramBuffer [3]string
	var counter uint8 = 0
	var previousOffset uint = 0
	files, err := GetFilesFromDir("/home/naamik/go/src/legacy/pressure679/GhostWriter/txts/")
	// var buffer []*string
	if err != nil { panic(err) }
	dataAsset, err := data.Asset("data/english.json")
	if err != nil { panic(err) }
	trainingData, err := sentences.LoadTraining(dataAsset)
	if err != nil { panic(err) }
	sentenceTokenizer := sentences.NewSentenceTokenizer(trainingData)
	mygraph.qaTriGrams = make(map[[3]string]uint)
	mygraph.qaNouns = make(map[string]uint)
	mygraph.qaNounToSentences = make(map[string][]uint)
	mygraph.qaNounToTriGram = make(map[string]uint)

	for fileNum, _ := range files {
		osFile, err := os.Open("/home/naamik/go/src/legacy/pressure679/GhostWriter/txts/" + files[fileNum])
		if err != nil { panic(err) }
		buffer, err := ioutil.ReadAll(osFile)
		if err != nil { panic(err) }
		previousOffset = uint(len(mygraph.Sentences) - 1)
		absSentences := sentenceTokenizer.Tokenize(string(buffer))
		for nssnum, _ := range absSentences {
			mygraph.Sentences = append(mygraph.Sentences, Sentence{Sentence: absSentences[nssnum].String()})
		}
		for sentenceNum, _ := range mygraph.Sentences[previousOffset + 1:] {
			words := fasttag.WordsToSlice(mygraph.Sentences[sentenceNum].Sentence)
			posTags := fasttag.BrillTagger(words)
			// mygraph.Sentences[sentenceNum].PosTags = posTags
			for wordNum, _ := range words {
				if string(posTags[wordNum][0]) != "N" { continue }
				mygraph.qaNounToSentences[words[wordNum]] = append(mygraph.qaNounToSentences[words[wordNum]], previousOffset + uint(sentenceNum))
				// trigram queue appendage
				if counter == 1 {
					trigramBuffer[2] = words[wordNum]
				} else if counter == 2 {
					trigramBuffer[1] = trigramBuffer[2]
					trigramBuffer[2] = words[wordNum]
				} else if counter == 3 {
					trigramBuffer[0] = trigramBuffer[1]
					trigramBuffer[1] = trigramBuffer[2]
					trigramBuffer[2] = words[wordNum]
					counter = 0
				}
				// init
				if len(mygraph.Nouns) == 0 {
					mygraph.qaNouns[words[wordNum]] = 0
					mygraph.Nouns = append(mygraph.Nouns, Noun{Noun:words[wordNum]})
					mygraph.Final.Nouns = append(mygraph.Final.Nouns, &mygraph.Nouns[0])
				}
				// append
				if _, exists := mygraph.qaNouns[words[wordNum]]; !exists {
					mygraph.Nouns = append(mygraph.Nouns, Noun{Noun:words[wordNum]})
					mygraph.Final.Nouns = append(mygraph.Final.Nouns, &mygraph.Nouns[len(mygraph.Nouns)-1])
					mygraph.qaNouns[words[wordNum]] = uint(len(mygraph.Nouns) - 1)
				} else { // increment occurence
					mygraph.Nouns[mygraph.qaNouns[words[wordNum]]].Occurence++
				}
				// wait for trigrambuffer
				if len(mygraph.Nouns) < 2 { continue }
				// append
				if _, exists := mygraph.qaTriGrams[trigramBuffer]; !exists {
					mygraph.TriGrams = append(mygraph.TriGrams, TriGram{Nouns: [3]*Noun{&mygraph.Nouns[mygraph.qaNouns[trigramBuffer[2]]], &mygraph.Nouns[mygraph.qaNouns[trigramBuffer[1]]], &mygraph.Nouns[mygraph.qaNouns[trigramBuffer[0]]]}})
					mygraph.Final.TriGrams = append(mygraph.Final.TriGrams, &mygraph.TriGrams[len(mygraph.TriGrams)-1])
					mygraph.qaTriGrams[trigramBuffer] = uint(len(mygraph.TriGrams) - 1)
					mygraph.qaNounToTriGram[words[wordNum]] = uint(len(mygraph.TriGrams) - 1)
				} else { // increment
					mygraph.TriGrams[mygraph.qaTriGrams[trigramBuffer]].Occurence++
				}
				counter++
			}
		}
		fmt.Println("Read", files[fileNum])
	}
	// Pruning?
	fmt.Println("\nnumber of sentences:", len(mygraph.Sentences))
	fmt.Println("number of nouns:", len(mygraph.Nouns))
	fmt.Println("number of trigrams:", len(mygraph.TriGrams))
	/* sumSentences := len(mygraph.Sentences)
	sumNouns := len(mygraph.Nouns)
	sumTriGrams := len(mygraph.TriGrams) */

	nounByOccurence := func(n1, n2 *Noun) bool {
		return n1.Occurence < n2.Occurence
	}
	trigramByOccurence := func(t1, t2 *TriGram) bool {
		return t1.Occurence < t2.Occurence
	}
	wg.Add(1)
	go func() {
		go ByNoun(nounByOccurence).Sort(mygraph.Final.Nouns)
		go ByTriGram(trigramByOccurence).Sort(mygraph.Final.TriGrams)
		wg.Done()
	}()
	wg.Wait()
	fmt.Println("\nDone sorting.")
	// As of now sentences are written more than once, and sentences are unstructured, both in what document they are from and where from the documents they are. - optimally a set of sentences should be organized by how they relate to one another (see literary tools).
	var strToFile string

	// var sentenceBreakPoint, nounBreakPoint map[*string]uint8 = make(map[*string]bool)
	var breakAll bool = false
	wg.Add(1)
	go func() {
		// trigram,noun,noun,noun
		for trigramPtrOffset, _ := range mygraph.Final.TriGrams {
			for trigramNounOffset, _ := range mygraph.Final.TriGrams[trigramPtrOffset].Nouns {

				for nounSentenceIndexNum, _ := range mygraph.qaNounToSentences[mygraph.Final.TriGrams[trigramPtrOffset].Nouns[trigramNounOffset].Noun] {
					uniques[&mygraph.Sentences[nounSentenceIndexNum].Sentence] = true
					strToFile += mygraph.Sentences[nounSentenceIndexNum].Sentence + "\n"
				}
			}
		}
		wg.Done()
	}()
	wg.Wait()
	go func() {
		final, err := os.Create("/home/naamik/go/src/legacy/pressure679/GhostWriter/sentenizer_final.txt")
		if err != nil { panic(err) }
		fmt.Fprintln(final, strToFile)
		err = final.Close()
		if err != nil { panic(err) }
	}()
	fmt.Println("Done.")
}
func GetFilesFromDir(directory string) (files []string, err error) {
osFileInfo, err := ioutil.ReadDir(directory); if err != nil { return nil, err }
for _, fileInfo := range osFileInfo { if !fileInfo.IsDir() { files = append(files, fileInfo.Name()) } }; return files, err
}
