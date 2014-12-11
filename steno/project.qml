import QtQml 2.0
import QtQml.Models 2.1
import QtQuick 2.3
import QtQuick.Controls 1.2
import QtQuick.Layouts 1.0
import QtQuick.Dialogs 1.2

Item {
    anchors.fill: parent
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

    Dialog {
        id: slurpProgessDlg
        width: 400
        height: 200
        title: "Slurping..."
        visible: (ctrl.slurpProgress.inFlight || ctrl.slurpProgress.errorMsg != "")
        contentItem: ColumnLayout {
            spacing: 4
            Text {
                Layout.fillWidth: true;
                visible: ctrl.slurpProgress.errorMsg != ""
                color: "#FF4444"
                font.bold: true
                wrapMode: Text.Wrap
                text: "ERROR: " + ctrl.slurpProgress.errorMsg }
            Text { text: "received "+ctrl.slurpProgress.totalCnt + " articles (" +ctrl.slurpProgress.newCnt + " new)" }
            BusyIndicator { running: ctrl.slurpProgress.inFlight }
        }
        standardButtons: StandardButton.Ok | StandardButton.Cancel
    }
}
