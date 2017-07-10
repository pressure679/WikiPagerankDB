# WikiPagerankDB
sample-boltdb and sample-mysql run, but the mwdumper.jar from the wikipedia source is a better solution to make a database (using mysql) from the wikipedia dump file(s).
tools.go contains an unfinished HMM tagger using the viterbi algorithm - question is if a nerual network is better suited than an in-case enumerator from probabilistic selection dependent on input.
TODO: implement the rake, tf-idf and pagerank (or another algorithm with lower memory consumption) algorithm to select suited articles and nouns from articles based on graph weights and nouns' popularity.

The user should input a desired article, and output should be somewhat like wikibooks - on the road to create an AI assistant for education/learning.

10/07-2017:
Added the rake function. it needs a final return argument in the NLP function, other than that the v.1 for producing new ghostwritten text based on the wikipaedia dump files just needs to utilize the given functions (arguments being article titles) and apply an A* algorithm or similar to traverse the articles' links. - Hmm, from here I planned to implement the naive bayes, tf-idf, and pagerank algorithms, where pagerank would definitely add an improvement. v.2 should communicate with a P2P network to share information (to save processing power) and make an interactive chat/learn bot (locally or through P2P with GraphQL) and/or matrix-like terminal with keyword-based coloring. (don't forget to Gob encode network traffic in for RPC and etc.) - There are also the translation option, rivescript, aiml for the keywords and sentence construction. And don't forget the perceptron based tagger if unknown writing styles or languages are included.
