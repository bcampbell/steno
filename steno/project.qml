import QtQml 2.0
import QtQml.Models 2.1
import QtQuick 2.3
import QtQuick.Controls 1.2
import QtQuick.Layouts 1.0
import QtQuick.Dialogs 1.0

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

}

