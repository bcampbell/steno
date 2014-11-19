import QtQml 2.0
import QtQml.Models 2.1
import QtQuick 2.3
import QtQuick.Controls 1.2
import QtQuick.Layouts 1.0

ApplicationWindow {
    id: window
    title: "Steno"
    visible: true

    menuBar: MenuBar {
        Menu {
            title: "File"
            MenuItem { text: "Open..." }
            MenuItem {
                text: "Close"
                shortcut: StandardKey.Close
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
                Text {
                    color: styleData.textColor
                    elide: styleData.elideMode
                    text: ctrl.art(styleData.row).article.headline
                }
            }
        }
        Component {
            id: pubDelegate
            Item {
                Text {
                    color: styleData.textColor
                    elide: styleData.elideMode
                    text: ctrl.art(styleData.row).pub
                }
            }
        }

        Component {
            id: publishedDelegate
            Item {
                Text {
                    color: styleData.textColor
                    elide: styleData.elideMode
                    text: ctrl.art(styleData.row).article.published
                }
            }
        }

        Component {
            id: urlDelegate
            Item {
                Text {
                    color: styleData.textColor
                    elide: styleData.elideMode
                    text: ctrl.art(styleData.row).article.canonicalURL
                }
            }
        }
        ColumnLayout {
            anchors.fill: parent
            TextInput {
                Layout.fillWidth: true
                text: "WIBBLE!"
            }
            TableView {
                Layout.fillHeight: true
                Layout.fillWidth: true
                id: mainView
                model: ctrl.len
                TableViewColumn{ role: "headline"  ; title: "Title" ; width: 100; delegate: headlineDelegate }
                TableViewColumn{ role: "pub"  ; title: "Publication" ; width: 100; delegate: pubDelegate }
                TableViewColumn{ role: "published"  ; title: "Published" ; width: 100; delegate: publishedDelegate }
                TableViewColumn{ role: "url" ; title: "URL" ; width: 200; delegate: urlDelegate  }
            }
    }
}
