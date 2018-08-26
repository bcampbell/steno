



Usage: steno-similar [OPTIONS] DBFILE1 DBFILE2
Options:
  -m float
    	match threshold (0=no matching, 1=all ngrams matched) (default 0.4)
  -n int
    	ngram size (default 3)
  -s int
    	ignore articles shorter than this this number of words (default 100)
  -v	verbose output

eg:

  $ steno-similar -m 0.5 -v poop1.db poop2.db


Indexing algorithm:

1) initialise the ngram table (to hold a list of doc IDs for each possible ngram)

2) For each document in the source:

    a) lowercase text, remove punctuation, remove stopwords, apply stemming
      eg:
      doc 1: "National Socialism is a bit silly."
      doc 2: "Nihilists! Fuck me. I mean, say what you want about the tenets of National Socialism, Dude, at least it's an ethos."

      becomes:
      doc 1: "nation social silly"
      doc 2: "nihili fuck mean tenet nation social dud ethos"

    b) split into ngrams
      eg (with ngramSize=2)
      doc 1: " nation social" "social silly"
      doc 2: "nihili fuck", "fuck mean", "mean tenet", "tenet nation", "nation social", "social dud" "dud least" "least ethos"

    c) go through each ngram and append the doc id to its slot in the ngram table.
       So, after indexing the two example docs above, the ngram table looks like this:

        "nation social"  => [1,2]
        "social silly" => [1]
        "nihili fuck"  => [2]
        "fuck mean"  => [2]
        "mean tenet"  => [2]
        "tenet nation"  => [2]
        "social dud"  => [2]
        "dud least"  => [2]
        "least ethos"  => [2]


Matching algorithm:

to match a document ("the query doc")against the index:

eg doc 3: "National socialism is fucking mean."

1) split up the document text into ngrams, as above.
   => doc 3: "nation social" "social fuck" "fuck mean"

2) get a list of _all_ the potential match docs which contain any of those ngrams
   =>
        "nation social"  => [1,2]
        "social fuck" => []  (empty - wasn't in original corpus)
        "fuck mean"  => [2]
        so the potential match docs are: [1,2]

   For each one we calculate a match factor:

   a) count how many of the query doc ngrams appear in this potential match doc
   => docid 1:  1 match
   => docid 2:  2 matches

   b) divide by the number of ngrams in the query doc to get match factor
   => docid 1: 1/3 = 0.33333...
   => docid 2: 2/3 = 0.66666...

   c) discard any documents below the match threshold

   d) perform a diff to compare the original text of the two documents
      (NOT the ngrams!) and display this in the HTML report







