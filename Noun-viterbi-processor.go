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
	"math"
	"strings"
)
type Sentence struct {
	Sentence string
	PosTags []string
	cfgs []uint8
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
	Nouns [3]*Noun
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
	qaTriGramToSentences map[*TriGram][]uint

	NounMaxOccurence uint
	NounMinOccurence uint
	NounSumOccurence uint
	NounMeanOccurence float64
	NounStdDev float64
	NounZScore float64

	TriGramMaxOccurence uint
	TriGramMinOccurence uint
	TriGramSumOccurence uint
	TriGramMeanOccurence float64
	TriGramStdDev float64
	TriGramZScore float64

	// Context-free grammar represented by POS-tags.
	cfg [][]string
	Final Pointers
	Print []*Sentence
}

func (graph Graph) LinearRegression() {
	graph.NounMinOccurence = 10000 // Default init value.
	// var sumWords uint
	for noun, _ := range graph.qaNounToSentences {
		if graph.NounMaxOccurence < uint(len(graph.qaNounToSentences[noun])) {
			graph.NounMaxOccurence = uint(len(graph.qaNounToSentences[noun]))
		}
		if graph.NounMinOccurence > uint(len(graph.qaNounToSentences[noun])) {
			graph.NounMinOccurence = uint(len(graph.qaNounToSentences[noun]))
		}
		graph.NounSumOccurence += uint(len(graph.qaNounToSentences[noun]))
	}
	graph.NounMeanOccurence = float64(graph.NounSumOccurence / uint(len(graph.Nouns)))
	graph.NounStdDev = math.Sqrt(float64(graph.NounSumOccurence)/graph.NounMeanOccurence)
	graph.NounZScore = graph.NounStdDev / float64(graph.NounMaxOccurence-graph.NounMinOccurence)
	graph.TriGramMinOccurence = 10000
	for trigramPtr, _ := range graph.qaTriGramToSentences {
		if len(graph.qaTriGramToSentences[trigramPtr]) == 0 {
			delete(graph.qaTriGramToSentences, trigramPtr)
			*trigramPtr = TriGram{}
			continue
		}
		if graph.TriGramMaxOccurence < uint(len(graph.qaTriGramToSentences[trigramPtr])) {
			graph.TriGramMaxOccurence = uint(len(graph.qaTriGramToSentences[trigramPtr]))
		}
		if graph.TriGramMinOccurence > uint(len(graph.qaTriGramToSentences[trigramPtr])) {
			graph.TriGramMinOccurence = uint(len(graph.qaTriGramToSentences[trigramPtr]))
		}
		graph.TriGramSumOccurence += uint(len(graph.qaTriGramToSentences[trigramPtr]))
	}
	graph.TriGramMeanOccurence = float64(graph.TriGramSumOccurence / uint(len(graph.qaTriGramToSentences)))
	graph.TriGramStdDev = math.Sqrt(float64(graph.TriGramSumOccurence) / graph.TriGramMeanOccurence)
	graph.TriGramZScore = graph.TriGramStdDev / float64(graph.TriGramMaxOccurence-graph.TriGramMinOccurence)
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
	mygraph.qaTriGramToSentences = make(map[*TriGram][]uint)
	mygraph.cfg = [][]string{
		[]string{"CD", "JJ", "NN"},
		[]string{"CD", "NN"},
		[]string{"CD", "RB", "JJ", "NN"},
		[]string{"DT", "CD", "JJ", "NN"},
		[]string{"DT", "CD", "NN"},
		[]string{"DT", "CD", "RB", "JJ", "NN"},
		[]string{"DT", "JJ", "CD", "JJ", "NN"},
		[]string{"DT", "JJ", "CD", "NN"},
		[]string{"DT", "JJ", "CD", "RB", "JJ", "NN"},
		[]string{"DT", "JJ", "JJ", "NN"},
		[]string{"DT", "JJ", "NN"},
		[]string{"DT", "JJ", "RB", "JJ", "NN"},
		[]string{"DT", "NN"},
		[]string{"DT", "RB", "JJ", "NN"},
		[]string{"DT", "RB", "VB", "NN"},
		[]string{"DT", "VB", "NN"},
		[]string{"JJ", "CD", "JJ", "NN"},
		[]string{"JJ", "CD", "NN"},
		[]string{"JJ", "CD", "RB", "JJ", "NN"},
		[]string{"JJ", "JJ", "NN"},
		[]string{"JJ", "NN"},
		[]string{"JJ", "RB", "JJ", "NN"},
		[]string{"NN"},
	}
	
	for fileNum, _ := range files {
		osFile, err := os.Open("/home/naamik/go/src/legacy/pressure679/GhostWriter/txts/" + files[fileNum])
		if err != nil { panic(err) }
		buffer, err := ioutil.ReadAll(osFile)
		if err != nil { panic(err) }
		previousOffset = uint(len(mygraph.Sentences) - 1)
		absSentences := sentenceTokenizer.Tokenize(string(buffer))
		for nssnum, _ := range absSentences {
			mygraph.Sentences = append(mygraph.Sentences, Sentence{Sentence: absSentences[nssnum].Text})
			words := fasttag.WordsToSlice(absSentences[nssnum].Text)
			posTags := fasttag.BrillTagger(words)
			mygraph.Sentences[nssnum + int(previousOffset)].PosTags = posTags
			for cfgnum, _ := range mygraph.cfg {
				for mgSscfgOffset := 0; mgSscfgOffset < len(mygraph.Sentences[nssnum + int(previousOffset)].PosTags); mgSscfgOffset++ {
					if mygraph.Sentences[nssnum + int(previousOffset)].PosTags[:len(mygraph.cfg[cfgnum])] == mygraph.cfg[cfgnum][:] {
						mygraph.Sentences[nssnum + int(previousOffset)].cfgs = append(mygraph.Sentences[nssnum + int(previousOffset)].cfgs, uint8(cfgnum)); lastOffset = len(mygraph.cfg[cfgnum]); cfgnum = 0
					}
				}
			}
		}
		for sentenceNum, _ := range mygraph.Sentences[previousOffset + 1:] {
			words := fasttag.WordsToSlice(mygraph.Sentences[sentenceNum].Sentence)
			for wordNum, _ := range words {
				if string(mygraph.Sentences[sentenceNum].PosTags[wordNum][0]) != "N" { continue }
				lowercaseWord := strings.ToLower(words[wordNum])
				mygraph.qaNounToSentences[lowercaseWord] = append(mygraph.qaNounToSentences[lowercaseWord], previousOffset + uint(sentenceNum))
				// trigram queue appendage
				if counter == 1 {
					trigramBuffer[2] = lowercaseWord
				} else if counter == 2 {
					trigramBuffer[1] = trigramBuffer[2]
					trigramBuffer[2] = lowercaseWord
				} else if counter == 3 {
					trigramBuffer[0] = trigramBuffer[1]
					trigramBuffer[1] = trigramBuffer[2]
					trigramBuffer[2] = lowercaseWord
					counter = 0
				}
				// init
				if len(mygraph.Nouns) == 0 {
					mygraph.qaNouns[lowercaseWord] = 0
					mygraph.Nouns = append(mygraph.Nouns, Noun{Noun:lowercaseWord})
					mygraph.Final.Nouns = append(mygraph.Final.Nouns, &mygraph.Nouns[0])
				}
				// append
				if _, exists := mygraph.qaNouns[lowercaseWord]; !exists {
					mygraph.Nouns = append(mygraph.Nouns, Noun{Noun:lowercaseWord})
					mygraph.Final.Nouns = append(mygraph.Final.Nouns, &mygraph.Nouns[len(mygraph.Nouns)-1])
					mygraph.qaNouns[lowercaseWord] = uint(len(mygraph.Nouns) - 1)
				} else { // increment occurence
					mygraph.Nouns[mygraph.qaNouns[lowercaseWord]].Occurence++
				}
				// wait for trigrambuffer
				if len(mygraph.Nouns) < 2 { continue }
				// append
				if _, exists := mygraph.qaTriGrams[trigramBuffer]; !exists {
					mygraph.TriGrams = append(mygraph.TriGrams, TriGram{Nouns: [3]*Noun{&mygraph.Nouns[mygraph.qaNouns[trigramBuffer[2]]], &mygraph.Nouns[mygraph.qaNouns[trigramBuffer[1]]], &mygraph.Nouns[mygraph.qaNouns[trigramBuffer[0]]]}})
					mygraph.Final.TriGrams = append(mygraph.Final.TriGrams, &mygraph.TriGrams[len(mygraph.TriGrams)-1])
					mygraph.qaTriGrams[trigramBuffer] = uint(len(mygraph.TriGrams) - 1)
					mygraph.qaNounToTriGram[lowercaseWord] = uint(len(mygraph.TriGrams) - 1)
					mygraph.qaTriGramToSentences[&mygraph.TriGrams[mygraph.qaTriGrams[trigramBuffer]]] = append(mygraph.qaTriGramToSentences[&mygraph.TriGrams[mygraph.qaTriGrams[trigramBuffer]]], previousOffset + uint(sentenceNum))
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

	mygraph.LinearRegression()
	nounByOccurence := func(n1, n2 *Noun) bool {
		return n1.Occurence < n2.Occurence
	}
	/* trigramByOccurence := func(t1, t2 *TriGram) bool {
		return t1.Occurence < t2.Occurence
	} */
	trigramByNounSumOccurence := func (t1, t2 *TriGram) bool {
		return t1.Nouns[0].Occurence + t1.Nouns[1].Occurence + t1.Nouns[2].Occurence < t2.Nouns[0].Occurence + t2.Nouns[1].Occurence + t2.Nouns[2].Occurence
	}
	wg.Add(1)
	go func() {
		go ByNoun(nounByOccurence).Sort(sort.Reverse(mygraph.Final.Nouns))
		// go ByTriGram(trigramByOccurence).Sort(mygraph.Final.TriGrams)
		go ByTriGram(trigramByNounSumOccurence).Sort(sort.Reverse(mygraph.Final.TriGrams))
		wg.Done()
	}()
	wg.Wait()
	fmt.Println("\nDone sorting.")
	// As of now sentences are written more than once, and sentences are unstructured, both in what document they are from and where from the documents they are. - optimally a set of sentences should be organized by how they relate to one another (see literary tools).
	// var strToFile string

	var uniqueSentences map[*string]bool = make(map[*string]bool)
	counter = 0
	var breakAll bool = false
	var breakerValue, uniqueCounter uint8 = uint8(mygraph.TriGramStdDev * mygraph.TriGramMeanOccurence), 0
	var debug, debugNum uint = 0, 1

	wg.Add(1)
	fmt.Println("Making sentences.")
	go func() {
		// trigram,noun,noun,noun,sentences
		for trigramPtrOffset, _ := range mygraph.Final.TriGrams {
			fmt.Println(*mygraph.Final.TriGrams[trigramPtrOffset].Occurence)
			for trigramNounOffset, _ := range mygraph.Final.TriGrams[trigramPtrOffset].Nouns {
				for nounSentenceIndexNum, _ := range mygraph.qaNounToSentences[mygraph.Final.TriGrams[trigramPtrOffset].Nouns[trigramNounOffset].Noun] {
					if uniqueSentences[&mygraph.Sentences[nounSentenceIndexNum].Sentence] { continue
					} else {
						uniqueSentences[&mygraph.Sentences[nounSentenceIndexNum].Sentence] = true
						fmt.Println(mygraph.Sentences[nounSentenceIndexNum].Sentence + "\n")
					}
				}
			}
		}
		wg.Done()
	}()
	wg.Wait()
	/* fmt.Println("Writing to file.")
	go func() {
		final, err := os.Create("/home/naamik/go/src/legacy/pressure679/GhostWriter/sentenizer_final.txt")
		if err != nil { panic(err) }
		fmt.Fprintln(final, strToFile)
		err = final.Close()
		if err != nil { panic(err) }
	}() */
	fmt.Println("Done.")
}
func GetFilesFromDir(directory string) (files []string, err error) {
osFileInfo, err := ioutil.ReadDir(directory); if err != nil { return nil, err }
for _, fileInfo := range osFileInfo { if !fileInfo.IsDir() { files = append(files, fileInfo.Name()) } }; return files, err
}
