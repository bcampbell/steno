# Query Syntax

## Basics

Single words are treated as simple terms. The results will include
articles containing those terms, eg:

    grapefruit

Terms are separated by spaces, so:

    navel orange

Will match any articles containing both "navel" AND "orange". Note that the
order and position of the two terms within a article is unimportant - as
long as the article contains both terms, it'll match.


To search for phrases - multiple words, matched in order - enclose them in
double quotes.

For example:

    "navel orange"

## Boolean Operators

* The `OR` operator returns articles matched by terms on either side of it.

        orange OR lemon

* `AND` returns articles which match the terms on *both* sides.
  It is the default operator, meaning that the following two queries are
  considered to be equivalent:

        orange AND "navel orange"
        orange "navel orange"


* `NOT` excludes articles that match the following term.
  For example, to match articles containing "orange" but not "paint":

        orange NOT paint



* `-` (minus sign) is the 'prohibit' operator. Prefixing a term with `-` will exclude matching articles.

  For example, to match articles containing "orange" but not "paint":

        orange -paint

* `+` (plus sign) is the 'required' operator. A term prefixed with `+` must exist in
  matching articles. You should never have to use this operator, as it's
  usually implied by default.

### Precedence

`-`, `+` and `NOT` take precedence over `AND`, which takes precedence over `OR`.

For example,

    lemon OR orange AND citrus OR grapefruit

is treated as:

    lemon OR (orange AND citrus) OR grapefruit



## Field Scoping

You can control which fields are matched by prefixing the name of a field separated by a colon.

Examples:

    author: Bob
    headline:"How to Make the Perfect Negroni"
    tags: (fruit OR paint)

If no field is specified, then all fields will be searched.

## Grouping

Parentheses can be used to group sub queries.
For example:

    content:(shaddock OR pomelo OR pamplemousse) AND (headline:fruit AND NOT headline:"fruit salad") AND tags:(greenish OR yellowish)

## Wildcards

Within an individual term, partial matches can be described using wildcard characters:

`?` to match any single character

`*` To match and sequence of zero or more characters

For example:

     qu?ck bro*


## Fuzziness

A fuzzy query is a term query that matches terms within a given
[Levenshtein distance](https://en.wikipedia.org/wiki/Levenshtein_distance).
The edit distance is the number of single-character edits (insertions, deletions or substitutions) allowed.

To specify a fuzzy query, use the tilde sign (`~`), optionally followed by the edit distance you'll accept.


For example,

    colour~1

to match "`colour`" or "`color`" (or "`zolour`" or "`colours`"... but not "`colors`", as the `~1` allows only a single character change).

If no number is specified, the default value is 2, so the following are equivalent:

    grapefruit~
    grapefruit~2



## Ranges

Inclusive ranges can be described with square braces (`[`, `]`) and `TO`, eg:

    retweets:[1 TO 5]
    published:[2010-01-01 TO 2010-01-31]

Dates must be in `YYYY-MM-DD` form.

Exclusive ranges are supported using curly braces (`{`,`}`). So, these are equivalent:

    retweets:{0 TO 10}
    retweets:[1 TO 9]



You can mix inclusive and exclusive endpoints, eg:

    published:[2010-01-01 TO 2011-01-01}
    favourites:{0 TO 16]
    retweets:[0 TO 256}

You can have open ranges by leaving off either endpoint:

    published:[2000-01-01 TO ]
    retweets:[TO 100}


NOTE: ranges currently work only on numeric and date fields.


## Relational Operators

You can perform numeric comparisons using the `>`, `>=`, `<`, and `<=` operators.
These are equivalent to using the above range syntax with unbounded ranges.

For example:

    retweets:>=100
    retweets:[ TO 100]

## More examples

All article in the express with `/sport/` in the url, excluding ones tagged as cruft:

    pub:express urls:/sport/ -tags:cruft

Articles about crime published in Feb 2010:

    tags:crime published:[2010-02-01 TO 2010-03-01}



TODO: more examples!



