# Appendix A - Data storage


The core of a Steno project is the `.db` file.
This is simply a [SQLite](http://sqlite.org/) database and can be viewed
and manipulated in any tool that handles SQLite databases.

Steno also maintains a parallel index directory (`.db.bleve`) which is
used to provide querying. It is implemented using
[Bleve](http://www.blevesearch.com). As the user modifies the database -
slurping in articles, tagging, deleting, etc - Steno keeps the `.db.bleve`
index updated.

It is safe to delete the `.db.bleve` index. When Steno opens a `.db`
project which has no corresponding `.db.bleve` directory, it will
pause to create it. This can take a while for a large project.












