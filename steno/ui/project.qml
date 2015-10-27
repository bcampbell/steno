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
                id: wibble
                Layout.fillHeight: true
                Layout.fillWidth: true
            }
            Item {
                id: artInfo
                Layout.minimumHeight: 100
                function showArt(art,highlightTerms) {
                   content.text = ctrl.renderContent(art);
                    console.log(content.text);
                }
                ScrollView {
                    anchors.fill: parent
                        anchors.margins: 16
                    contentItem: Text {
                        id: content 
        //                width: artInfo.width
                        width: 600
                        text: ""


                        wrapMode: Text.WordWrap
                        textFormat: Text.RichText
                    }
                
                }
            }
        }
    }


    Dialog {
        id: progressDlg
        width: 500
        height: 100
        title: ctrl.progress.title
        visible: (ctrl.progress.inFlight || ctrl.progress.errorMsg != "")
        contentItem: ColumnLayout {
            anchors.fill: parent
        anchors.margins: 12
            spacing: 4
            ProgressBar {
                Layout.fillWidth: true;
                visible: ctrl.progress.inFlight
                value: ctrl.progress.completedCnt
                minimumValue: 0
                maximumValue: ctrl.progress.expectedCnt
                indeterminate: (ctrl.progress.expectedCnt==0) ? true:false
            }
            Text {
                Layout.fillWidth: true;
                visible: ctrl.progress.errorMsg != ""
                color: "#FF4444"
                font.bold: true
                wrapMode: Text.Wrap
                text: "ERROR: " + ctrl.progress.errorMsg }
            Text { text: ctrl.progress.statusMsg }
            //BusyIndicator { running: ctrl.progress.inFlight }
        }
        standardButtons: StandardButton.Ok | StandardButton.Cancel
    }

}
