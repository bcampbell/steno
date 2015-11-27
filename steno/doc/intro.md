% Steno Users Manual
% Ben Campbell
% Nov 2015


# Introduction

TODO: flesh out intro properly, screenshots etc

## General Workflow

1. Open a new project (menu: `File->New`)

    Steno projects are stored in a `.db` database file.
    There is no need to explicitly save work as changes are directly
    applied to the database as you go along.

2. Slurp articles from a server (menu: `tools->Slurp articles from server`)

    'Slurping' is the process of downloading articles from a remote server
    which is collecting articles.

    Typically, you'd slurp down articles for a day, or a range of days.

    You can perform multiple slurps, adding articles to the same project.
    Slurping can be safely repeated - Steno will ignore articles already
    in a project and only add articles it doesn't already have.

3. Use queries to filter articles of interest

    The query syntax is detailed below, but in general it works much
    like any of the text search boxes you'd encounter on the web.

    The matching articles are listed in the main window. You can click on an article
    to have it's content appear in the lower pane.

4. Apply tags

    Tagging is useful for classifying articles according to whatever
    criteria you're interested in.

    Tags are single words, lower case.
    Beyond that, you are free to make up your own tagging conventions.

    Currently-matching articles can be selected by clicking on them.
    Multiple articles can be selected by holding down `SHIFT` for range
    selection or `CTRL` to toggle individual articles.

    `CTRL-A` will select *all* currently-matching articles.

    Once selected, articles can be tagged by entering the name of the tag and
    pressing the "Add Tags" button.

    You can remove tags and delete articles in the same way.

5. Bundle common tasks into scripts

    Once you've established a way of working - querying, tagging, refining
    your data - you can automate repetitive tasks by creating scripts.

    See the below section on [scripting](#scripting) for details.


## Article fields

Articles in steno have the following fields:

*	`headline`
*	`urls` - the URL(s) of the article. There can be more than one.
*	`content` - the full text of the article
*	`pub` - source publication
*	`byline` - the authors of the article
*	`published` - the date of publication
*	`keywords` - any keywords or tags the publication has applied to the article
*	`section` - which section within the publication the article appears (if known)
*	`tags` - the user-applied tags
*	`retweets`  (tweets only)
*	`favourites`  (tweets only)
*	`links` - links within the tweet (tweets only)

All of these fields may be used in queries.

Steno has two separate view modes, accessible under the menu `View->Mode`.
One mode is designed for articles, the other for tweets.

The only difference between modes is in which field columns are displayed
in the matching-articles list. The mode has no effect upon the underlying data.




