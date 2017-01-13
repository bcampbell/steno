import QtQuick.Controls 1.2
import QtQuick.Layouts 1.0
import QtQuick.Dialogs 1.2

import "helper.js" as Helper
Dialog {
    function pad (num, size) {
        var s = num+"";
        while (s.length < size) s = "0" + s;
        return s;
    }

    width: 500
    height: 300

    standardButtons: StandardButton.Ok | StandardButton.Cancel
    title: "Build a fasttext model from tagged articles"
    ColumnLayout {
        anchors.fill: parent

        GridLayout {
            columns: 2
            rowSpacing: 5
            columnSpacing: 5

            Label { text:"Epoch" }
            ColumnLayout {
                SpinBox {
                    id: epochField
                    minimumValue:1
                    maximumValue: 10000
                    value: 100
                    //Layout.fillWidth: true;
                }
            }
            Label {}
            Label {
                text: "Higher values mean longer/better training.\nExperimentation required.\n(200 seems ok)"
                font.italic: true
                //wrapMode:wrap
               // lineHeight: 3
            }
        }
    }

    FileDialog {
        id: fileDialog
        title: "Create model file"
        selectExisting: false
        onAccepted: {
            var outFile = Helper.filePathFromURL(fileDialog.fileUrl);
            ctrl.train(outFile,epochField.value);
        }

        onRejected: {
        }
    }

    onAccepted: {
        fileDialog.open()
    }
}

