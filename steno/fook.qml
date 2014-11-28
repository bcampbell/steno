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



    SplitView {
        anchors.fill: parent
        orientation: Qt.Horizontal

        SplitView {
            Layout.fillHeight: true
            Layout.fillWidth: true
            orientation: Qt.Vertical
            Query {
                Layout.fillHeight: true
                Layout.fillWidth: true
            }
            Item {
                id: artInfo
                Layout.minimumHeight: 100
                function showArt(art) {
                   content.text = art.content 
                }
                ScrollView {
                    anchors.fill: parent
                    Text {
                        id: content 
        //                width: artInfo.width
                        //anchors.fill: parent
                        //anchors.margins: 16
                        text: ""


                        wrapMode: Text.WordWrap
                        textFormat: Text.StyledText
                    }
                
                }
            }
        }

        HelpPane {
            id: helpPane

/*
            Layout.fillHeight: true
*/
            Layout.fillWidth: true
            helpText: ctrl.helpText
        }
    }
}




