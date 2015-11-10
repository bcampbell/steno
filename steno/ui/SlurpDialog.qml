//import QtQml 2.0
//import QtQml.Models 2.1
//import QtQuick 2.3
import QtQuick.Controls 1.2
import QtQuick.Layouts 1.0
import QtQuick.Dialogs 1.2
//import QtQuick.Window 2.2

Dialog {
    function pad (num, size) {
        var s = num+"";
        while (s.length < size) s = "0" + s;
        return s;
    }

    width: 320
    height: 400

    standardButtons: StandardButton.Ok | StandardButton.Cancel
    title: "Slurp articles from server"
    ColumnLayout {
        anchors.fill: parent
        Label { text:"Source" }
        ComboBox {
            id: slurpSource
            model: {
                var names = [];
                for( var i=0; i<app.slurpSourcesLen; ++i ) {
                    names.push(app.getSlurpSourceName(i));
                }
                return names;
            }
        }
        Label { text:"Pick day(s)" }
        FancyCal {
            id: dayPicker
            maxDays: 14
         }
    }


    onAccepted: {
        var fromD = dayPicker.dateStart;
        var toD = dayPicker.dateEnd;
        var fromStr = pad(fromD.getFullYear(),4) + '-' + pad(fromD.getMonth()+1,2) + '-' + pad(fromD.getDate(),2);
        var toStr = pad(toD.getFullYear(),4) + '-' + pad(toD.getMonth()+1,2) + '-' + pad(toD.getDate(),2);
        app.current().slurp( slurpSource.currentText, fromStr, toStr);
    }

}

