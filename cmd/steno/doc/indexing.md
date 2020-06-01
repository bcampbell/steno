# Text Indexing Details

This section goes into more detail on how the text indexing operates.
You can probably skip it, but it provides some insight which can help
in diagnosing issues with querying.


During indexing, article data is 'cooked' down into individual terms which are
then entered into the index. 

The index is like the index you'd find in the back of a book. It's an
easy-to-scan list of terms, and beside each term is a list of the articles
which contain that term.

For most fields, this is merely a case of splitting the text up into words,
and lower-casing them.
But some fields have more elaborate rules.

The same cooking rules applied to the articles are also applied to queries
before a search is performed.
This is important because terms either match exactly, or not at all.
You cannot match partial terms (ignoring wildcard operators and fuzzy operators).
So `tory` will *not* match `history`. It's `tory` or nothing.

## `headline`, `content` and `byline` fields

When indexing the `headline`, `content` and `byline` fields,
some extra processing is applied. 

This is best illustrated by an example: let's take the following text:

    None of Bob's connections!

1. The text is split up into words and unwanted punctuation is removed:

        None of Bob's connections

2. Possessive suffixes are removed:

        None of Bob connections

3. lowercasing

        none of bob connections

4. stopword removal (we'll use `_` to indicate a removed word)

        none _ bob connections

    See [Appendix B](#appendix-b---list-of-stopwords) for a full list of stopwords.

5. stemming (using the porter stemming algorithm):

        none _ bob connect


So the final phrase would be indexed as: `none _ bob connect`.

When querying, the search text is passed through the same process. Some example queries:

    content:"Bob connected"   -> "bob connect" (MATCH!)
    content:"none the Bob"    -> "none _ bob" (MATCH!)
    content:"none wibble Bob" -> "none wibble bob" (NO MATCH)
    content:"None of Bob's"   -> "none _ bob" (MATCH)



## `urls` and `links` fields


Some special processing is also applied to URLs.
A URL (or query) like:

    http://example.com/foo/bar/moon-made-of-cheese

is split up into:

    http : / / example . com / foo / bar / moon - made - of - cheese

so these should all match fine:

    moon-made-of-cheese
    /foo/
    example.com

but these will not:

    www.example.com     (no "www .")
    moon-made-of-chee   (no "chee" in index. Only "cheese")


## `published` field

Internally, dates are stored as numbers (the number of nanoseconds since Jan 1, 1970)

This means that you _have_ to use ranges to match the `published` field.
so if you want everything on May 25th 2001, do:

    published:[2001-05-25 TO 2001-05-25]

and NOT:

    published:2001-05-25

Yes, this *is* sucky, but for now we're stuck with it.




## More about stemming

Mostly, the stemming is concerned with snipping off suffixes to get at a standardised root word.
eg `connection`, `connected`, `connecting`, `connects` all stem to `connect`.

Note that the stemming mostly works OK, but will fall prey to the bizarreness of English language some times. For example, I think it'll treat `business` as a form of `busy` (both will end up as `busi`, I think. Which isn't tooooo unreasonable, but likely not what you want)







