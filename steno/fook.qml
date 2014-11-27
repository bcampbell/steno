import QtQml 2.0
import QtQml.Models 2.1
import QtQuick 2.3
import QtQuick.Controls 1.2
import QtQuick.Layouts 1.0
import QtQuick.Dialogs 1.0

/* by convention:
    id
    property declarations
    signal declarations
    JavaScript functions
    object properties
    child objects
    states
    transitions
*/


ApplicationWindow {
    id: window
    width: 900
    height: 600

    title: "Steno"
    visible: true

    FileDialog {
        id: openDialog
        title: "Please choose a file"

        function toLocalFile(f) {
            return f.toString().replace(/^file:\/\//, "");
        }


        onAccepted: {
            console.log("You chose: " + toLocalFile(openDialog.fileUrl))
            ctrl.loadDB(toLocalFile(openDialog.fileUrl))
             
        }
        onRejected: {
            console.log("Canceled")
        }
        //Component.onCompleted: visible = true
    }

    Action {
        id: openFile
        //iconSource: "images/fileopen.png"
        text: "Open..."
        shortcut: StandardKey.Open
        onTriggered: openDialog.open()
    }
    Action {
        id: close
        //iconSource: "images/fileopen.png"
        text: "Close"
        shortcut: StandardKey.Close
        onTriggered: ctrl.close()
    }

    menuBar: MenuBar {
        Menu {
            title: "File"
            MenuItem {
                action: openFile
            }
            MenuItem {
                action: close
            }
        }
    }

    statusBar: StatusBar {
        RowLayout {
            Label { text: "Read Only" }
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


    SplitView {
        anchors.fill: parent
        orientation: Qt.Vertical
        ColumnLayout {
            //anchors.fill: parent
            Layout.fillHeight: true
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
            GridView {
                Layout.fillWidth: true
                Layout.fillHeight: true
                height: 10
                cellWidth: 150
                cellHeight: 20
                model: ctrl.facetLen
                delegate: Row {
                        Rectangle {
                anchors.fill: parent
                            color: "white"
                            border.width:4
                            border.color:"black"
                            Row {
                            spacing: 4
                            Text { text: ctrl.facet(index).txt }
                            Text { text: ctrl.facet(index).cnt }
                            }
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

        Item {
            id: artInfo
            Layout.minimumHeight: 100
            function showArt(art) {
               content.text = art.content 
            }
            Text {
                id: content 
//                width: artInfo.width
                anchors.fill: parent
                anchors.margins: 16
                text: ""


                wrapMode: Text.WordWrap
                textFormat: Text.StyledText
            }
            
        }
    }

}




