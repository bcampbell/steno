# Query Syntax

Steno provides a basic text query language to perform searches for articles
matching desired criteria.

## Basics

The simplest search is a single word, a term. All matching documents will contain
that term.

Multiple terms can separated by spaces, so:

    navel orange

Will match any articles containing both `navel` AND `orange`. The
order and position of the two terms within a article is unimportant.
As long as the article contains both terms, it will match.

To search for phrases - multiple words, matched in order - enclose them in
double quotes.

For example:

    "navel orange"

## Boolean Operators

* The `OR` operator matches articles matched by terms on either side of it.

        orange OR lemon

* `AND` matches articles which contain *both* terms.
  It is the default operator, so the following two queries are
  considered to be equivalent:

        orange AND "navel orange"
        orange "navel orange"


* `NOT` excludes articles which match a term.
    For example, to match articles containing `orange` but not `paint`:

        orange NOT paint


* `-` (minus sign) is the 'prohibit' operator. Prefixing a term with `-` will exclude matching articles.
    For example, to match articles containing `orange` but not `paint`:

        orange -paint

    For most intents, `-` is equivalent to using `NOT`.


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

You can control which fields are matched by prefixing the name of a field, separated by a colon.

Examples:

    byline: "Bob Smith"
    headline:"How to Make the Perfect Negroni"
    tags: (fruit OR paint)

If no field is specified, then all fields will be searched.

Note that the field prefix only applies to the term immediately following
the colon. So, for example, this query is probably incorrect:

    byline: Bob Smith

It would match any articles with `Bob` in the byline and `smith` in any field.



## Grouping

Parentheses can be used to group sub queries.
For example:

    content:(pomelo OR pamplemousse) AND tags:(fruit -cruft)

## Wildcards

Within an individual term, partial matches can be described using wildcard characters:

`?` to match any single character

`*` To match and sequence of zero or more characters

For example:

     qu?ck bro*


## Fuzziness

A fuzzy query is a query that matches terms within a given
[Levenshtein distance](https://en.wikipedia.org/wiki/Levenshtein_distance).
This is the number of single-character edits (insertions, deletions or
substitutions) allowed between two matching terms.

To specify a fuzzy query, use the tilde sign (`~`), optionally followed by the distance you'll accept.

For example,

    colour~1

to match "`colour`" or "`color`" (or "`zolour`" or "`colours`"), but
not "`colors`", as the `~1` allows only a single character change.

If the number is omitted the default value is 2. So the following are equivalent:

    grapefruit~
    grapefruit~2



## Ranges

Inclusive ranges can be described with square braces (`[`, `]`) and `TO`.
For example:

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
These are equivalent to using the above range syntax with open ranges.

For example:

    retweets:>=100
    retweets:[ TO 100]

## Examples

All article in the express with `/sport/` in the url, excluding ones tagged as cruft:

    pub:express urls:/sport/ -tags:cruft

Articles about crime published during February 2010:

    tags:crime published:[2010-02-01 TO 2010-03-01}



