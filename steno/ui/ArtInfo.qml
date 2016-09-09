import QtQml 2.0
//import QtQml.Models 2.1
import QtQuick 2.3
import QtQuick.Controls 1.2
import QtQuick.Layouts 1.0
//import QtQuick.Dialogs 1.2




Item {
    id: artInfo
    property var art: null
    property var hlTerms: []

    function showArt(a,highlightTerms) {
//       content.text = ctrl.renderContent(art,highlightTerms);
//        headline.text = art.headline
        hlTerms = highlightTerms;
        art = a; 
    }

    SplitView {
        handleDelegate: SplitHandle { }
        anchors.fill: parent
        ColumnLayout {
            Layout.minimumWidth: 40
            Layout.margins: 16
            Text {
                text: art ? art.headline : ""
                font.bold: true
                font.pixelSize: 24
                elide: Text.ElideRight
                maximumLineCount: 2
                wrapMode: Text.Wrap
            }

            Label { text: "CanonicalURL:" }
            Text {
                Layout.leftMargin: 16
                text: art ? art.canonicalURL : ""
            }
            Label { text: "Published:" }
            Text {
                Layout.leftMargin: 16
                text: art ? art.published : ""
            }
            Label { text: "URLs:" }
            Repeater {
                model: art ? art.numURLs(): 0;
                Text {
                    Layout.leftMargin: 16
                    text: art.getURL(index)
                }
            }

        }
        TextArea {
            Layout.minimumWidth: 40
            Layout.margins: 0
        //    id: content 
            Layout.fillWidth: true
            Layout.fillHeight: true
            //width: 600
            text: art===null? "" : ctrl.renderContent(art,hlTerms);
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

