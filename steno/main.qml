import QtQml 2.0
import QtQml.Models 2.1
import QtQuick 2.3
import QtQuick.Controls 1.2
import QtQuick.Layouts 1.0
import QtQuick.Dialogs 1.2
import QtQuick.Window 2.2

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
            console.log("You chose: " + toLocalFile(newDialog.fileUrl))
            app.openProject(toLocalFile(newDialog.fileUrl))
             
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
        enabled: !app.hasCurrent
    }

    Action {
        id: newAction
        //iconSource: "images/fileopen.png"
        text: "New..."
        shortcut: StandardKey.New
        onTriggered: newDialog.open()
        enabled: !app.hasCurrent
    }
    Action {
        id: closeAction
        //iconSource: "images/fileopen.png"
        text: "Close"
        shortcut: StandardKey.Close
        onTriggered: app.closeProject()
        enabled: app.hasCurrent
    }
    Action {
        id: quitAction
        //iconSource: "images/fileopen.png"
        text: "Quit"
        shortcut: StandardKey.Quit
        onTriggered: app.quit()
    }
    Action {
        id: slurpAction
        //iconSource: "images/fileopen.png"
        text: "Slurp articles from server..."
        onTriggered: slurpDlg.open()    //app.current().slurp()
        enabled: app.hasCurrent
    }
    Action {
        id: helpAction
        //iconSource: "images/fileopen.png"
        text: "Help..."
        shortcut: StandardKey.HelpContents
        onTriggered: helpWindow.visible = !helpWindow.visible
    }

    menuBar: MenuBar {
        Menu {
            title: "File"
            MenuItem { action: openAction }
            MenuItem { action: newAction }
            MenuItem { action: closeAction }
            MenuItem { action: quitAction }
            MenuItem { action: slurpAction }
        }
        Menu {
            title: "Help"
            MenuItem { action: helpAction }
        }
    }
    Item {
        anchors.fill: parent
        objectName: "mainSpace"
    }

    statusBar: StatusBar {
        RowLayout {
            Label { text: app.errorMsg }
        }
    }


    Window {
        id: helpWindow
        title: "Help"
        width: 400
        height: 500
        ScrollView {
            anchors.fill: parent
            Text {
                width: 400
                //text: helpText
                wrapMode: Text.Wrap
                textFormat: Text.RichText
                //width: parent.width
                text: ""+ app.helpText
                Layout.maximumWidth: 400
            }
        }

    }



    Dialog {
        id: slurpDlg
        title: "Slurp articles from server"

        contentItem: Column {
            spacing: 4
            Label { text:"Pick day" }
            Calendar {
                id: dayPicker
                onDoubleClicked: slurpDlg.click(StandardButton.Ok)
             }
        }
        standardButtons: StandardButton.Ok | StandardButton.Cancel
        onAccepted: app.current().slurp(
            dayPicker.selectedDate.toISOString().slice(0,10),
            dayPicker.selectedDate.toISOString().slice(0,10))
    }
}





