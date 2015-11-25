# Text Indexing Details

During indexing, article data is 'cooked' down into individual terms which are
then entered into the index. 

For most fields, this is merely a case of splitting the text up into words,
and lower-casing them.
But some fields have more elaborate rules.

The same rules are applied to queries before a search is performed.
This is important because terms either match exactly, or not at all.
You cannot match partial terms (unless you're using the above wildcard operators).
So "tory" will NOT match "history". It's "tory" or nothing.

## `headline`, `content` and `byline` fields

When indexing `headline`, `content` and `byline` fields have
some extra processing is applied. 

Best illustrated by an example: let's take the following text:

    `None of Bob's connections!`

1. The text is split up into words:

        None of Bob's connections

2. Possessive suffixes are removed:

        None of Bob connections

3. lowercasing

        none of bob connections

4. stopword removal (we'll use `_` to indicate a removed word)

        none _ bob connections

5. stemming (using the porter stemming algorithm):

        none _ bob connect


So the final phrase would be indexed as: `none _ bob connect`.

When querying, the search text is passed through the same process. Some example queries:

    content:"Bob connected"   -> "bob connect" (MATCH!)
    content:"none the Bob"    -> "none _ bob" (MATCH!)
    content:"none wibble Bob" -> "none wibble bob" (NO MATCH)
    content:"None of Bob's"   -> "none _ bob" (MATCH)



## `urls` field


Some processing is also applied to URLs.
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

Internally, dates are stored as numbers (the number of nanoseconds since jan 1 1970 I think :- )

This means that you _have_ to use ranges to match the "published" field.
so if you want everything on may 25th 2001, do:

    published:[2001-05-25 TO 2001-05-25]

and NOT:

    published:2001-05-25

Yes, I know this is sucky, but for now we're stuck with it.




## More about stemming

Mostly, the stemming is concerned with snipping off suffixes to get at a standardised root word.
eg "connection", "connected", "connecting", "connects" all stem to "connect".

Note that the stemming mostly works OK, but will fall prey to the bizarreness of English language some times. For example, I think it'll treat "business" as a form of "busy" (both will end up as "busi", I think. Which isn't tooooo unreasonable, but likely not what you want)


## List of stopwords

The default list of stopwords seems pretty sensible (see below).
To me, the only one that looked like it might cause trouble is "against".

The full list is:

    i
    me
    my
    myself
    we
    our
    ours
    ourselves
    you
    your
    yours
    yourself
    yourselves
    he
    him
    his
    himself
    she
    her
    hers
    herself
    it
    its
    itself
    they
    them
    their
    theirs
    themselves
    what
    which
    who
    whom
    this
    that
    these
    those
    am
    is
    are
    was
    were
    be
    been
    being
    have
    has
    had
    having
    do
    does
    did
    doing
    would
    should
    could
    ought
    i'm
    you're
    he's
    she's
    it's
    we're
    they're
    i've
    you've
    we've
    they've
    i'd
    you'd
    he'd
    she'd
    we'd
    they'd
    i'll
    you'll
    he'll
    she'll
    we'll
    they'll
    isn't
    aren't
    wasn't
    weren't
    hasn't
    haven't
    hadn't
    doesn't
    don't
    didn't
    won't
    wouldn't
    shan't
    shouldn't
    can't
    cannot
    couldn't
    mustn't
    let's
    that's
    who's
    what's
    here's
    there's
    when's
    where's
    why's
    how's
    a
    an
    the
    and
    but
    if
    or
    because
    as
    until
    while
    of
    at
    by
    for
    with
    about
    against
    between
    into
    through
    during
    before
    after
    above
    below
    to
    from
    up
    down
    in
    out
    on
    off
    over
    under
    again
    further
    then
    once
    here
    there
    when
    where
    why
    how
    all
    any
    both
    each
    few
    more
    most
    other
    some
    such
    no
    nor
    not
    only
    own
    same
    so
    than
    too
    very






