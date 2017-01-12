import QtQuick 2.2
import QtQuick.Controls 1.2
import QtQuick.Layouts 1.0
import QtQuick.Dialogs 1.2

import "helper.js" as Helper
Dialog {
    width: 500
    height: 300

    standardButtons: StandardButton.Ok | StandardButton.Cancel
    title: "Autotag articles using fasttext model"
    ColumnLayout {
        anchors.fill: parent

        Label { text:"Probability threshold" }
        RowLayout {
            Slider {
                id: thresholdField
                value: 0.1
                minimumValue: 0.0
                maximumValue: 1.0
                stepSize: 0.01
            }
            TextField {
                validator: DoubleValidator {bottom: 0.0; top: 1.0; decimals: 2 }
                text: (thresholdField.value).toFixed(2);
                onEditingFinished: {
                    thresholdField.value = Number(text);
                }
            }
        }
        Label {
            text: "Articles will only have a tag applied if\nit's estimated probability is greater than this value"
            font.italic: true
            //wrapMode:wrap
           // lineHeight: 3
        }
    }

    FileDialog {
        id: fileDialog
        title: "Pick model file"
        selectExisting: true
        onAccepted: {
            var modelFile = Helper.filePathFromURL(fileDialog.fileUrl);
            ctrl.autoTag(modelFile,thresholdField.value);
        }

        onRejected: {
        }
    }

    onAccepted: {
        fileDialog.open()
    }
}

