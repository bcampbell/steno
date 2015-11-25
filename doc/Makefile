SRCS := intro.md syntax.md scripting.md indexing.md

HTML = /tmp/steno.html
PDF = /tmp/steno.pdf


all: $(HTML) $(PDF)

$(HTML): $(SRCS)
	pandoc $^ --toc --standalone -o $@

$(PDF): $(SRCS)
	pandoc $^ --toc --standalone -o $@

