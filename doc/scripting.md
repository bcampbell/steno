# Scripting

Scripts are text files found in the `scripts` directory. They should be named with
a `.txt` extension (or `.csv` - see below).

You can create subdirectories to better organise groups of scripts.
Each subdirectory will appear as a separate tab in the script-picking
dialog box.

## Script format

Scripts are line-based - each command starts on a new line.
Blank lines are ignored, so it's fine to use them to improve readabilty.

Commands are of the form:

`<query> => <operation>`

where `<operation>` is one of:

  * `TAG <tag1> ... <tagN>`

    to add tags to matching articles

  * `UNTAG <tag1> ... <tagN>`

    to remove tags from matching articles

  * `DELETE`

    to delete matching articles


The commands are run in the order they appear in the script.


## Comments

The hash symbol (`#`) denotes the start of a comment. The rest of the
line is simply ignored.

Comments can be placed on their own line, or can appear after a command.

If the first line of the script is a comment, it is used as the description
for the script.


## Example script - `fruit.txt`

    # identify important fruit-related new articles

    orange OR lemon OR grapefruit => tag fruit citrus
    raspberry OR strawberry => TAG fruit berry

    # uncomment this line to handle tomatoes
    #tomato => UNTAG fruit


## Alternate script format (.csv)

There is also a simpler script format, in the form of a .csv file.
These are comma-separated-value files, a somewhat standardised file format
which most spreadsheet applications can use.

Steno will automatically determine the script type by the file extension
(.txt or .csv).

In the alternate format, the first line determines the meaning of each column.
On this line each cell is either the name of a field to search (eg "headline",
"content" etc), or "TAG".

Each subsequent line forms a single query. The contents of each cell are
matched against their field, and any matching articles are tagged with any
values found in "TAG" columns.

Complex queries are not supported: the search terms are just matched against
their fields and nothing else.

The value in "TAG" columns is optional - so different lines can apply
different numbers of tags if desired.

## Example alternative script - `fruit.csv`

    content, TAG, TAG
    orange, fruit, citrus
    lemon, fruit, citrus
    grapefruit, fruit, citrus
    raspberry, fruit, berry
    strawberry, fruit, berry
    plum,,fruit

(note the empty TAG column on plum - we only want to apply one tag
for that query)



