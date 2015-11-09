import QtQml 2.0
//import QtQml.Models 2.1
import QtQuick 2.3
import QtQuick.Controls 1.2
import QtQuick.Layouts 1.0
//import QtQuick.Dialogs 1.2

Item {
    id: artInfo
    function showArt(art,highlightTerms) {
       content.text = ctrl.renderContent(art,highlightTerms);
        headline.text = art.headline
    }
    ColumnLayout {
        anchors.fill: parent
        anchors.margins: 4
        Text {
            id: headline
            font.bold: true
            font.pixelSize: 12
            wrapMode: Text.Wrap
        }
        TextArea {
            id: content 
            Layout.preferredWidth: 600
            Layout.fillHeight: true
            width: 600
            //height: artInfo.height
            text: ""
            readOnly: true

            wrapMode: Text.WordWrap
            textFormat: Text.RichText

            // TODO: update to QtQuick.Controls 1.3 (QT5.5?) which has
            // proper context-menu support in TextArea (see "menu" member)
            Menu {
                id: editMenu
                title: "Edit"

                MenuItem {
                    text: "Copy"
                    shortcut: "Ctrl+C"
                    onTriggered: { content.copy(); }
                }
            }
        }
    }
    // TODO: ditch MouseArea when possible! Stops hyperlinks working.
    MouseArea {
        acceptedButtons: Qt.RightButton
        propagateComposedEvents: true
        visible: true;
        anchors.fill: parent
            onClicked: { editMenu.popup(); }
    }


}

