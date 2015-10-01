function filePathFromURL(url) {
    var f = url.toString();
    if (!f.match(/^file:\/\//i)) {
        return f;   // not a file:// url
    }
    f = f.replace(/^file:\/\//i, "");

    // evil hack for windows "/c:/foo/bar/..." need to strip leading /
    if (f.match(/^\/[a-z]:/i)) {
        f = f.slice(1);
    }

    return f;
}


// mark up links in plain text
function markupLinks(txt) {
    var url_match = /https?:\S+/g;
    var out = txt.replace(url_match,'<a href="$&">$&</a>');
    return out;
}


