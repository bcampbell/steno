import QtQml 2.0
import QtQml.Models 2.1
import QtQuick 2.3
import QtQuick.Controls 1.2
import QtQuick.Layouts 1.0
import QtQuick.Dialogs 1.0


// This is the main bit, where the query and results and tools are shown.

Item {
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
                text: "" + ctrl.len + " matching articles (of " + ctrl.totalArts + ")"
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
                model: ctrl.facetLen
                delegate: Rectangle {
                    width: childrenRect.width + 8
                    height: childrenRect.height + 8
                    border.width: 1
                    border.color: Qt.darker(color,2)
                    radius: 4
                    color: "#eeeeff"
                    Text { x:4; y:4; text: ctrl.facet(index).txt + ": " +ctrl.facet(index).cnt }
                }
            }
        }
    Component {
        id: headlineDelegate
        Item {
            clip: true
            Text {
                anchors.fill: parent
                color: styleData.textColor
                elide: Text.ElideRight
                text: ctrl.art(styleData.row).headline

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
                text: ctrl.art(styleData.row).pub
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
                text: ctrl.art(styleData.row).published
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
                text: asLink(ctrl.art(styleData.row).canonicalURL)
                onLinkActivated: Qt.openUrlExternally(link)
                    
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
                text: ctrl.art(styleData.row).tagsString()
            }
        }
    }

        TableView {
            id: artList
            Layout.fillHeight: true
            Layout.fillWidth: true
            objectName: "artlist"
            selectionMode: SelectionMode.ExtendedSelection
            model: ctrl.len
            function selectedArts() {
                var sel = [];
                selection.forEach( function(rowIndex) { sel.push(rowIndex)} )
                return sel
            }

            onClicked: artInfo.showArt(ctrl.art(row))
            TableViewColumn{ role: "headline"  ; title: "headline" ; width: 400; delegate: headlineDelegate }
            TableViewColumn{ role: "pub"  ; title: "pub" ; width: 100; delegate: pubDelegate }
            TableViewColumn{ role: "published"  ; title: "published" ; width: 100; delegate: publishedDelegate }
            TableViewColumn{ role: "tags" ; title: "tags" ; width: 100; delegate: tagsDelegate  }
            TableViewColumn{ role: "url" ; title: "url" ; width: 400; delegate: urlDelegate  }
        }
    }
}
