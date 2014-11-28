import QtQml 2.0
import QtQml.Models 2.1
import QtQuick 2.3
import QtQuick.Controls 1.2
import QtQuick.Layouts 1.0
import QtQuick.Dialogs 1.0

ScrollView {
    property string helpText
    id: pane
    Text {
        //text: helpText
        wrapMode: Text.Wrap
        textFormat: Text.RichText
        //width: parent.width
        text: ""+ helpText
        Layout.maximumWidth: 400
    }
}

