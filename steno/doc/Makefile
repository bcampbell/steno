CSS = style.css
SRCS := intro.md syntax.md indexing.md scripting.md appendix_a.md appendix_b.md

HTML = steno.html
PDF = steno.pdf

OUTPUTS = $(HTML) $(PDF)


all: $(OUTPUTS)

clean:
	rm -f $(OUTPUTS)

.PHONY: all clean

$(HTML): $(SRCS) $(CSS)
	pandoc $(SRCS) --toc -t html5 --css $(CSS) --standalone -o $@

$(PDF): $(SRCS)
	pandoc $^ --toc -N --standalone -o $@


