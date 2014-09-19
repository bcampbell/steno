function toolup(container,targ) {
    var termRe = /\b([.a-zA-Z]+):\s*(("[^"]*")|('[^']*')|(\[(.*?)])|(\S*))/g;

    var dateRe = /^(\d{4}-\d{2}-\d{2})$/;
    var dateRangeRe = /^\[\s*(\d{4}-\d{2}-\d{2})\s+TO\s+([-0-9]{0,10})\s*\]$/;

    var rethink = function() {
        var cntTerm=0, cntRange=0, cntPub=0;

        // clear out the container
        while(container.hasChildNodes()) {
            container.removeChild(container.lastChild);
        }

        // for each text pattern that matches, create a widget for twiddling it...
        while((m=termRe.exec(targ.value)) !==null) {

            var field = m[1];
            var val = m[2];
            var txt = m[0]
            
            createTwiddler(field,val,m.index,m.index+m[0].length);
        }

        var helpers = document.importNode(document.querySelector("#tmpl-helpers").content,true);

        var addPub = helpers.querySelector(".helper-addpub");
        if(container.querySelector('.twiddler [name="pub"]') !== null) {
            addPub.style.display = "none";
        } else {
            addPub.addEventListener("click",function() {
                targ.value = (targ.value + ' pub:').trim();
                rethink();
            });
        }

        var addRange = helpers.querySelector(".helper-addrange");
        if(container.querySelector('.twiddler [name="from"]') !== null) {
            addRange.style.display = "none";
        } else {
            addRange.addEventListener("click",function() {
                targ.value = (targ.value + ' published:[2014-04-23 TO 2014-05-31]').trim();
                rethink();
            });
        }

        container.appendChild(helpers);

    return;
        var btn;
        btn = container.querySelector(".twiddler-addterm");
        if( cntTerm >0 ) {
            btn.style.display='none';
        } else {
            btn.addEventListener("click",function() {
                targ.value = targ.value + ' headline:"grapefruit"';
                rethink();
                return false;
            });

        }


        btn = container.querySelector(".twiddler-addrange");
        if( cntRange >0 ) {
            btn.style.display='none';
        } else {
            btn.addEventListener("click",function() {
                targ.value = targ.value + ' published:[2014-04-23 TO 2014-05-31]';
                rethink();
                return false;
            });

        }

        btn = container.querySelector(".twiddler-addpub");
        if( cntPub >0 ) {
            btn.style.display='none';
        } else {
            btn.addEventListener("click",function() {
                targ.value = targ.value + ' pub:';
                rethink();
                return false;
            });
        }
    };


    function createTwiddler(field,val,txtBegin,txtEnd) {
        console.log("createTwiddler",field,val);


        // decide which kind of widget to use...
        var widget = null;
        if( field == "pub") {
            widget = publicationEditor(val);
        } else {
            var dm = dateRangeRe.exec(val);
            if(dm!==null) {
                // it's a date range
                widget = dateRangeEditor(dm[1],dm[2]);
            } else {
                dm = dateRe.exec(val);
                if(dm!==null) {
                    widget = dateEditor(dm[1]);
                }
            }
        }

        if(widget==null) {
            return;
        }


        var tmpl = document.getElementById('tmpl-twiddler');
        var t = tmpl.content.cloneNode(true);

        t = t.querySelector('.twiddler');   // easier to deal with real nodes than DocumentFragments
        var inpClose = t.querySelector('.close');
        var placeholder = t.querySelector(".placeholder");

        t.replaceChild(widget.frag, placeholder);

        inpClose.addEventListener("click", function() {
            // delete the content
            var old = targ.value;
            targ.value = old.slice(0,txtBegin) + old.slice(txtEnd);
            rethink();
        });

        t.addEventListener("change", function() {
            var val = widget.val();
            var old = targ.value;
            targ.value = old.slice(0,txtBegin) + field + ":" + val + old.slice(txtEnd);
            rethink();
        });

        container.appendChild(t);
    }

    function dateRangeEditor(fromDay,toDay) {
        var tmpl = document.getElementById('tmpl-edit-daterange');
        var t = tmpl.content.cloneNode(true);
        var inpFrom = t.querySelector('[name="from"]');
        var inpTo = t.querySelector('[name="to"]');

        inpFrom.value = fromDay;
        inpTo.value = toDay;

        return {
            val: function() { return "["+inpFrom.value + " TO " + inpTo.value + "]"; },
            frag: t
        };
    }


    function dateEditor(day) {
        var tmpl = document.getElementById('tmpl-edit-date');
        var t = tmpl.content.cloneNode(true);
        var inpDay = t.querySelector('[name="day"]');

        inpDay.value = day;

        return {
            val: function() { return inpDay.value; },
            frag: t
        };
    }

    function publicationEditor(pub) {
        var tmpl = document.getElementById('tmpl-edit-publication');
        var t = tmpl.content.cloneNode(true);
        var inp = t.querySelector('[name="pub"]');

        inp.value = pub;

        return {
            val: function() { return inp.value; },
            frag: t
        };
    }


    targ.addEventListener( "input", rethink);
    rethink();
}

