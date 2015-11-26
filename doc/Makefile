CSS = style.css
SRCS := intro.md syntax.md indexing.md scripting.md

HTML = /tmp/steno.html
PDF = /tmp/steno.pdf

OUTPUTS = $(HTML) $(PDF)


all: $(OUTPUTS)

clean:
	rm -f $(OUTPUTS)

.PHONY: all clean

$(HTML): $(SRCS) $(CSS)
	pandoc $(SRCS) --toc -t html5 --css $(CSS) --standalone -o $@

$(PDF): $(SRCS)
	pandoc $^ --toc -N --standalone -o $@


