# Note

As of now I have:
- swapped between keeping an index in RAM, or writing it to a more persistent storage
- compute markov chains of sentences
- tag the content (english text) for their POS (Part-Of-Speech) as defined by a brill tagger (see github.com/mvryan/fasttag)
- train a markov chain on the penn treebank by Python's NLTK data (see http://www.nltk.org/nltk_data/ #18)
- extract nouns and count their frequency in a document
- pagerank some articles based on links
- I have also swapped the heading of the project from mysql, github.com/boltdb/bolt, to gzipping, zipping, and pure text.

Do what you want with it.

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
