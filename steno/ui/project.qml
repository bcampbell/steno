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
                   content.text = art.formatContent(highlightTerms);
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
                        textFormat: Text.StyledText
                    }
                
                }
            }
        }
    }


    Dialog {
        id: progressDlg
        width: 400
        height: 200
        title: ctrl.progress.title
        visible: (ctrl.progress.inFlight || ctrl.progress.errorMsg != "")
        contentItem: ColumnLayout {
            spacing: 4
            Text {
                Layout.fillWidth: true;
                visible: ctrl.progress.errorMsg != ""
                color: "#FF4444"
                font.bold: true
                wrapMode: Text.Wrap
                text: "ERROR: " + ctrl.progress.errorMsg }
            Text { text: ctrl.progress.statusMsg }
            BusyIndicator { running: ctrl.progress.inFlight }
        }
        standardButtons: StandardButton.Ok | StandardButton.Cancel
    }

}
