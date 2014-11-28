import QtQml 2.0
import QtQml.Models 2.1
import QtQuick 2.3
import QtQuick.Controls 1.2
import QtQuick.Layouts 1.0
import QtQuick.Dialogs 1.0

Item {
    property string helpText
    id: pane

    width: 400
    Column {
        anchors.fill: parent
        Text {
            text: helpText
            wrapMode: Text.NoWrap
            textFormat: Text.RichText
        }
/*
        Button {
            text: "Close"
            onClicked: {pane.width=0}
        }
*/
    }
}

