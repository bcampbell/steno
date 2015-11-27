import QtQml 2.0
import QtQml.Models 2.1
import QtQuick 2.3
import QtQuick.Controls 1.2
import QtQuick.Layouts 1.0
import QtQuick.Dialogs 1.2
import QtQuick.Window 2.2


import "helper.js" as Helper

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
        onAccepted: {
            var f = Helper.filePathFromURL(openDialog.fileUrl);
            app.openProject(f)
        }
        onRejected: {
        }
        //Component.onCompleted: visible = true
    }
    FileDialog {
        id: newDialog
        title: "Create new project"
        nameFilters: [ "Steno database files (*.db)", "All files (*)" ]
        selectExisting: false

        onAccepted: {
            var f = Helper.filePathFromURL(newDialog.fileUrl);
            app.openProject(f)
             
        }
        onRejected: {
        }
        //Component.onCompleted: visible = true
    }

    FileDialog {
        id: exportOverallsDialog
        title: "Export overall summary"
        nameFilters: [ "CSV files files (*.csv)", "All files (*)" ]
        selectExisting: false

        onAccepted: {
            var f = Helper.filePathFromURL(exportOverallsDialog.fileUrl);
            app.current().exportOveralls(f)
        }
        onRejected: {
        }
        //Component.onCompleted: visible = true
    }

    FileDialog {
        id: exportCSVDialog
        title: "Export current matches as CSV"
        nameFilters: [ "CSV files files (*.csv)", "All files (*)" ]
        selectExisting: false

        onAccepted: {
            var f = Helper.filePathFromURL(exportCSVDialog.fileUrl);
            app.current().exportCSV(f)
        }
        onRejected: {
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
        id: runScriptAction
        text: "Run script..."
        onTriggered: {
                app.refreshScripts();
                pickScriptDlg.open();
        }
        enabled: app.hasCurrent
    }
    Action {
        id: embiggenShortLinksAction
        text: "Embiggen shortlinks..."
        onTriggered: {
                app.current().embiggenShortlinks();
        }
        enabled: app.hasCurrent
    }
    Action {
        id: tagRetweetsAction
        text: "Tag all retweets..."
        onTriggered: {
                app.current().tagRetweets();
        }
        enabled: app.hasCurrent
    }
    Action {
        id: exportOverallsAction
        //iconSource: "images/fileopen.png"
        text: "Export overall summary csv..."
        onTriggered: exportOverallsDialog.open()
        enabled: app.hasCurrent && app.current().results.len > 0
    }
    Action {
        id: exportCSVAction
        //iconSource: "images/fileopen.png"
        text: "Export matching articles to .csv..."
        onTriggered: exportCSVDialog.open()
        enabled: app.hasCurrent && app.current().results.len > 0
    }
    Action {
        id: trainAction
        text: "Train"
        onTriggered: app.current().train()
        enabled: app.hasCurrent
    }
    Action {
        id: classifyAction
        text: "Classify"
        onTriggered: app.current().classify()
        enabled: app.hasCurrent
    }
    Action {
        id: helpAction
        //iconSource: "images/fileopen.png"
        text: "Help..."
        shortcut: StandardKey.HelpContents
        onTriggered: app.openManual()
    }

    ExclusiveGroup {
        id: viewModeGroup
        Action {
            id: articleModeAction
            text: "Articles"
            checkable: true
            checked: true
            onToggled: { if(app.hasCurrent && checked) { app.current().setViewMode("article") }; }
            enabled: app.hasCurrent
        }

        Action {
            id: tweetModeAction
            text: "Tweets"
            checkable: true
            enabled: app.hasCurrent
            onToggled: { if( app.hasCurrent && checked) { app.current().setViewMode("tweet") }; }
        }
    }

    ExclusiveGroup {
        id: dateFmtGroup
        Action {
            id: dateFmt1
            text: "Mon 02/01/2006"
            checkable: true
            checked: true
            onToggled: { if(app.hasCurrent && checked) { app.current().setDateFmt(text) }; }
            enabled: app.hasCurrent
        }
        Action {
            id: dateFmt2
            text: "2006-01-02"
            checkable: true
            onToggled: { if(app.hasCurrent && checked) { app.current().setDateFmt(text) }; }
            enabled: app.hasCurrent
        }
        Action {
            id: dateFmt3
            text: "2006-01-02 15:04"
            checkable: true
            onToggled: { if(app.hasCurrent && checked) { app.current().setDateFmt(text) }; }
            enabled: app.hasCurrent
        }
        Action {
            id: dateFmt4
            text: "2006-01-02T15:04:05Z07:00"
            checkable: true
            onToggled: { if(app.hasCurrent && checked) { app.current().setDateFmt(text) }; }
            enabled: app.hasCurrent
        }
    }

    menuBar: MenuBar {
        Menu {
            title: "File"
            MenuItem { action: openAction }
            MenuItem { action: newAction }
            MenuItem { action: closeAction }
            MenuItem { action: quitAction }
        }
        Menu {
            title: "Tools"
            MenuItem { action: runScriptAction }
            //MenuItem { action: trainAction }
            //MenuItem { action: classifyAction }
            MenuItem { action: embiggenShortLinksAction }
            MenuItem { action: tagRetweetsAction }
            MenuSeparator { }
            MenuItem { action: exportCSVAction }
            MenuItem { action: exportOverallsAction }
            MenuSeparator { }
            MenuItem { action: slurpAction }
        }
        Menu {
            title: "View"
            Menu {
                title: "Mode"
                MenuItem { action: tweetModeAction }
                MenuItem { action: articleModeAction }
            }
            Menu {
                title: "Date format"
                MenuItem { action: dateFmt1 }
                MenuItem { action: dateFmt2 }
                MenuItem { action: dateFmt3 }
                MenuItem { action: dateFmt4 }
            }
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



    SlurpDialog {
        id: slurpDlg;
    }



    Dialog {
        id: pickScriptDlg
        title: "Pick script to run..."
        width: 650
        height: 400
        contentItem: ColumnLayout {
            spacing: 8
            TabView {
                Layout.fillWidth: true
                Layout.fillHeight: true

                // create a tab per category....
                Repeater {
                    model: app.scriptCategoriesLen
                    Tab {
                        title: app.getScriptCategory(index)

                        TableView {
                            id: scriptList
                            model: ListModel {
                                // filter scripts by category
                                Component.onCompleted:
                                {
                                    var cat = app.getScriptCategory(index)
                                    for (var i = 0; i < app.scriptsLen; i++)
                                    {
                                        var s = app.getScript(i)
                                        if(s.category == cat) {

                                            append( { idx:i, name: s.name, desc: s.desc } )
                                        }
                                    }
                                }
                            }

                            TableViewColumn {
                                role: "name"
                                title: "Name"
                                width: 200
                            }
                            TableViewColumn {
                                role: "desc"
                                title: "Description"
                                width: 400
                            }
                            onDoubleClicked: {
                                var idx = model.get(currentRow).idx;
                                app.current().runScript(idx);
                                pickScriptDlg.close();
                                
                            }
                        }
                    }
                }
            }
        }
    }
}





