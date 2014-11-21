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
        onTriggered: window.close()
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
                text: ctrl.art(styleData.row).article.headline
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
                text: ctrl.art(styleData.row).article.published
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
                text: asLink(ctrl.art(styleData.row).article.canonicalURL)
                    
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
                // TODO: no reason we can't validate the query properly
                onEditingFinished: ctrl.setQuery(text)
            }
            Text {
                text: "" + ctrl.len + " matching articles (of " + ctrl.totalArts + ")"
            }


            TableView {
                Layout.fillHeight: true
                Layout.fillWidth: true
                objectName: "artlist"
                selectionMode: SelectionMode.ExtendedSelection
                model: ctrl.len

                onClicked: artInfo.showArt(ctrl.art(row))
                TableViewColumn{ role: "headline"  ; title: "headline" ; width: 400; delegate: headlineDelegate }
                TableViewColumn{ role: "pub"  ; title: "pub" ; width: 100; delegate: pubDelegate }
                TableViewColumn{ role: "published"  ; title: "published" ; width: 100; delegate: publishedDelegate }
                TableViewColumn{ role: "url" ; title: "url" ; width: 400; delegate: urlDelegate  }
            }
        }

        Item {
            id: artInfo
            Layout.minimumHeight: 100
            function showArt(art) {
               content.text = art.article.content 
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




