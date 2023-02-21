func (graph *Graph) LinearRegression() {
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
	graph.NounMeanOccurence = float64(float64(graph.NounSumOccurence) / float64(len(graph.Nouns)))
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
	graph.TriGramMeanOccurence = float64(float64(graph.TriGramSumOccurence) / float64(len(graph.qaTriGramToSentences)))
	graph.TriGramStdDev = math.Sqrt(float64(graph.TriGramSumOccurence) / graph.TriGramMeanOccurence)
	graph.TriGramZScore = graph.TriGramStdDev / float64(graph.TriGramMaxOccurence-graph.TriGramMinOccurence)
}
