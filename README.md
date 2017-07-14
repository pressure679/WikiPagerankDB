# WikiPagerankDB

## sample-boltdb.go
On Linux use the commands:

go get github.com/pressure679/WikiPagerankDB

go build $GOPATH/src/github.com/pressure679/WikiPagerankDB/sample-boltdb.go

cd $GOPATH/src/github.com/pressure679/WikiPagerankDB/

mkdir articles

cd articles

And depending on your machine (PC I recomment downloading the ~200MB-~2GB files, a 60+GB RAM/VRAM machine the single data dump is fine). Type the url https://dumps.wikimedia.org/enwiki/ into a webbrowser's URL input line, and a list of wikipedia data dump distributions should come up. Select the one you want.

wget https://dumps.wikimedia.org/enwiki/latest/enwiki-latest-pages-articles...

cd ..

./sample-boltdb

This is going to take some days, almost a week, with ~4GB RAM, ~2.4 GHz CPU, but do not worry about shutting down the PC, you can move the already used wikipedia dump files away from the directory and let the program continue from where it left off.

## sample-mysql.go
For this I recommend the mwdumper, see https://www.mediawiki.org/wiki/Manual:MWDumper
This I tested for ~20 minutes before I saw it only added ~50 articles, then I figured I used a base MySQL schema and decidede to make my own index of wikipedia and just use the raw xml files. (although MySQL or a custom compression or encoding would be nice).
For this install mysql and add a database named wikidb, then add your mysql username and password into the <username> and <password> field on line 71.

### Description of samples
sample-boltdb and sample-mysql run, and the same functionality of these are in tools.go but with the modification of making an index instead of the boltdb and/or mysql functionality.

## tools.go
This contains a HMM tagger using the viterbi algorithm, although the ratio of occurrence of each markov chain (trigram of sentences) is not added (ratio is for perceptron usage).

TODO: the NLP function and TagArticle function, then a P2P network with a server of existing peers.

the NLP function should produce a text given an argument of given nouns and articles.
The TagArticle should take in a list of articles as defined by the return argument #1 in the ReadWikiXML function, then utilize the return arguments from the LoadMarkovChain function to use a perceptron based tagger, perceptron natwork for each set of possible sentence structures.
- The possible sentence structures is yet to be added, but this is already read in from the Penn Treebank corpus
A Brill based POS-tagger is also possible, and a viterbi based POS-tagger might just require less computation but more memory.

10/07-2017:
Added the rake function. it needs a final return argument in the NLP function, other than that the v.1 for producing new ghostwritten text based on the wikipaedia dump files just needs to utilize the given functions (arguments being article titles) and apply an A* algorithm or similar to traverse the articles' links.

- Hmm, from here I planned to implement the naive bayes, tf-idf, and pagerank algorithms, where pagerank would definitely add an improvement. v.2 should communicate with a P2P network to share information (to save processing power) and make an interactive chat/learn bot (locally or through P2P with GraphQL) and/or matrix-like terminal with keyword-based coloring. (don't forget to Gob encode network traffic in for RPC and etc.) - There are also the translation option, rivescript, aiml for the keywords and sentence construction. And don't forget the perceptron based tagger if unknown writing styles or languages are included.
The naive bayes tagger would base off of the given list of nouns, which should be given as argument, and rank other nouns in the articles' corpus'.
The tf-idf algorithm would be like the naive bayes functionality, but it will use the articles' corpus' as arguments and rank the nouns in each articles for a total noun-based pagerank.
The pagerank function will require max 100GB RAM, which can be ~halfed by implementing a Gob encoding to the articles, and yet less memory can be used by modifying the pagerank algorithm to make forests of neighboring articles. (a data structuring issue here)
