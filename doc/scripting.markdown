## Scripting

Scripts are text files found in the `scripts` directory. They should be named with
a `.txt` extension.

Blank lines are ignored, so it's fine to use them to improve readabilty.

### Commands

Scripts are line-based - each command starts on a new line.

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


### Comments

The hash symbol (`#`) denotes the start of a comment. The rest of the
line is simply ignored.

Comments can be placed on their own line, or can appear after a command.

If the first line of the script is a comment, it is used as the description
for the script.


### Example script - `fruit.txt`

    # identify important fruit-related new articles

    orange OR lemon OR grapefruit => tag fruit citrus
    raspberry OR strawberry => TAG fruit berry

    # uncomment this line to handle tomatoes
    #tomato => UNTAG fruit

