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
        title: "Please choose project to open"
        nameFilters: [ "Steno database files (*.db)", "All files (*)" ]
        function toLocalFile(f) {
            return f.toString().replace(/^file:\/\//, "");
        }


        onAccepted: {
            console.log("You chose: " + toLocalFile(openDialog.fileUrl))
            app.openProject(toLocalFile(openDialog.fileUrl))
             
        }
        onRejected: {
            console.log("Canceled")
        }
        //Component.onCompleted: visible = true
    }
    FileDialog {
        id: newDialog
        title: "Create new project"
        nameFilters: [ "Steno database files (*.db)", "All files (*)" ]
        selectExisting: false
        function toLocalFile(f) {
            return f.toString().replace(/^file:\/\//, "");
        }

        onAccepted: {
            console.log("You chose: " + toLocalFile(openDialog.fileUrl))
            app.openProject(toLocalFile(openDialog.fileUrl))
             
        }
        onRejected: {
            console.log("Canceled")
        }
        //Component.onCompleted: visible = true
    }

    Action {
        id: openAction
        //iconSource: "images/fileopen.png"
        text: "Open..."
        shortcut: StandardKey.Open
        onTriggered: openDialog.open()
    }

    Action {
        id: newAction
        //iconSource: "images/fileopen.png"
        text: "New..."
        shortcut: StandardKey.New
        onTriggered: newDialog.open()
    }
    Action {
        id: closeAction
        //iconSource: "images/fileopen.png"
        text: "Close"
        shortcut: StandardKey.Close
        onTriggered: app.closeProject()
    }
    Action {
        id: quitAction
        //iconSource: "images/fileopen.png"
        text: "Quit"
        shortcut: StandardKey.Quit
        onTriggered: app.quit()
    }

    menuBar: MenuBar {
        Menu {
            title: "File"
            MenuItem { action: openAction }
            MenuItem { action: newAction }
/*            MenuItem { action: closeAction } */
            MenuItem { action: quitAction }
        }
    }
    Item {
        anchors.fill: parent
        objectName: "mainSpace"
    }

    statusBar: StatusBar {
        RowLayout {
            Label { text: "Read Only" }
        }
    }



}





