import QtQml 2.0
import QtQml.Models 2.1
import QtQuick 2.3
import QtQuick.Controls 1.2
import QtQuick.Layouts 1.0
import QtQuick.Dialogs 1.0


// This is the main bit, where the query and results and tools are shown.

Item {

    function currentSel() {
        var foo = -1;
        artList.selection.forEach(function(i) { foo=i; });
        return foo;
    }

    function findNext(needle) {
        var i = currentSel();
        if(i==-1) {
            i=0;
        } else {
            i=i+1;
        }
        i = ctrl.results.findForward(i,needle);
        if (i<0) {
            // wrap?
        } else {
            artList.selection.clear();
            artList.selection.select(i);
            artList.positionViewAtRow(i, ListView.Center);
        }
    }
    function findPrevious(needle) {
        var i = currentSel()-1;
        if(i<0) {
            i=0;
        }
        i = ctrl.results.findReverse(i,needle);
        if (i<0) {
            // wrap?
        } else {
            artList.selection.clear();
            artList.selection.select(i);
            artList.positionViewAtRow(i, ListView.Center);
        }
    }

    Action {
        id: findAction
        text: "Find..."
        shortcut: StandardKey.Find
        onTriggered: findText.focus=true
    }
    Action {
        id: findNextAction
        text: "Find next"
        shortcut: StandardKey.FindNext
        onTriggered: findNext(findText.text)
        enabled: findText.text!=""
    }
    Action {
        id: findPreviousAction
        text: "Find previous"
        shortcut: StandardKey.FindPrevious
        onTriggered: findPrevious(findText.text)
        enabled:  findText.text!=""
    }

    ColumnLayout {
        anchors.fill: parent
    //    Layout.fillHeight: true
        TextField {
            objectName: "query"
            Layout.fillWidth: true
            text: ""
            placeholderText: "filter"
            // TODO: no reason we can't validate the query properly
            onEditingFinished: ctrl.setQuery(text)
        }
        RowLayout {
            Text {
                text: "" + ctrl.results.len + " matching articles (of " + ctrl.totalArts + ")"
            }
            TextField {
                id: tagEntry
                objectName: "tagEntry"
                text: ""
                placeholderText: "tag"
            }
            Button {
                id: buttonAddTag
                enabled: artList.selection.count > 0 && tagEntry.text!=""
                text: "add tag"
                onClicked: ctrl.addTag(artList.selectedArts(), tagEntry.text)
            }
            Button {
                id: buttonRemoveTag
                enabled: artList.selection.count > 0 && tagEntry.text!=""
                text: "remove tag"
                onClicked: ctrl.removeTag(artList.selectedArts(), tagEntry.text)
            }
            Button {
                id: buttonDeleteArts
                enabled: artList.selection.count > 0
                text: "delete"
                onClicked: ctrl.deleteArticles(artList.selectedArts())
            }
            Text {
                text: "" + artList.selection.count + " articles selected"
            }
        }

        // show facets
        Flow {
            Layout.fillWidth: true
           // columns: width/150
            spacing: 4
            Repeater {
                model: ctrl.results.facetLen
                delegate: Rectangle {
                    width: childrenRect.width + 8
                    height: childrenRect.height + 8
                    border.width: 1
                    border.color: Qt.darker(color,2)
                    radius: 4
                    color: "#eeeeff"
                    Text { x:4; y:4; text: ctrl.results.facet(index).txt + ": " +ctrl.results.facet(index).cnt }
                }
            }
        }


        // the results display
        TableView {

        Component {
            id: headlineDelegate
            Item {
                clip: true
                Text {
                    anchors.fill: parent
                    color: styleData.textColor
                    elide: Text.ElideRight
                    text: ctrl.results.art(styleData.row).headline

                }
            }
        }

        Component {
            id: pubDelegate
            Item {
                clip: true
                Text {
                    anchors.fill: parent
                    color: styleData.textColor
                    elide: styleData.elideMode
                    text: ctrl.results.art(styleData.row).pub
                }
            }
        }

        Component {
            id: publishedDelegate
            Item {
                clip: true


                Text {
                    anchors.fill: parent
                    color: styleData.textColor
                    elide: styleData.elideMode
                    text: ctrl.results.art(styleData.row).published
                }
            }
        }

        Component {
            id: urlDelegate
            Item {
                function asLink(s) {
                    return '<a href="'+s+'">'+s+'</a>';
                }
                clip: true
                Text {
                    anchors.fill: parent
                    color: styleData.textColor
                    elide: Text.ElideRight
                    text: asLink(ctrl.results.art(styleData.row).canonicalURL)
                    onLinkActivated: Qt.openUrlExternally(link)
                        
                }
            }
        }

        Component {
            id: sectionDelegate
            Item {
                clip: true
                Text {
                    anchors.fill: parent
                    color: styleData.textColor
                    elide: styleData.elideMode
                    text: ctrl.results.art(styleData.row).section
                }
            }
        }

        Component {
            id: tagsDelegate
            Item {
                clip: true
                Text {
                    anchors.fill: parent
                    color: styleData.textColor
                    elide: Text.ElideRight
                    text: ctrl.results.art(styleData.row).tagsString()
                }
            }
        }

            id: artList
            Layout.fillHeight: true
            Layout.fillWidth: true
            objectName: "artlist"
            selectionMode: SelectionMode.ExtendedSelection
            sortIndicatorVisible: true
            sortIndicatorColumn: ctrl.sortColumn
            sortIndicatorOrder: ctrl.sortOrder
            model: ctrl.results.len
            function selectedArts() {
                var sel = [];
                selection.forEach( function(rowIndex) { sel.push(rowIndex)} )
                return sel
            }

            onClicked: artInfo.showArt(ctrl.results.art(row))
            onSortIndicatorColumnChanged: ctrl.applySorting(sortIndicatorColumn, sortIndicatorOrder)
            onSortIndicatorOrderChanged: ctrl.applySorting(sortIndicatorColumn, sortIndicatorOrder)
            TableViewColumn{ role: "headline"  ; title: "headline" ; width: 400; delegate: headlineDelegate }
            TableViewColumn{ role: "pub"  ; title: "pub" ; width: 100; delegate: pubDelegate }
            TableViewColumn{ role: "section"  ; title: "section" ; width: 100; delegate: sectionDelegate }
            TableViewColumn{ role: "published"  ; title: "published" ; width: 100; delegate: publishedDelegate }
            TableViewColumn{ role: "tags" ; title: "tags" ; width: 100; delegate: tagsDelegate  }
            TableViewColumn{ role: "url" ; title: "url" ; width: 400; delegate: urlDelegate  }
        }
        Row {
            id: findBar
            TextField {
                id: findText
                placeholderText: "find"
                onEditingFinished: {
                    artList.selection.clear();
                    findNext(text);
                }
            }
            Button {
                text:"Next"
                action: findNextAction
            }
            Button {
                text:"Prev"
                action: findPreviousAction
            }
        }
    }
}
