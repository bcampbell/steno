
import QtQuick 2.3

import QtQuick.Controls 1.2
import QtQuick.Controls.Styles 1.2

// FancyCal is a calendar widget which lets you drag out a range of dates

Calendar {
    id: cal
    property date dateStart: new Date((new Date()).setHours(0, 0, 0, 0))
    property date dateEnd: new Date((new Date()).setHours(0, 0, 0, 0))
    property int maxDays: 0

//    anchors.centerIn: parent
    style: CalendarStyle {
        dayDelegate: Rectangle {
            id: delegate;
            color: (cal.dateStart<=styleData.date && styleData.date<=cal.dateEnd)  ? "#55c" :"white"
            Label {
                text: styleData.date.getDate()
                font.pixelSize: 12
                anchors.centerIn: parent
                color: styleData.visibleMonth ? 
                    ((cal.dateStart<=styleData.date && styleData.date<=cal.dateEnd)  ? "white" :"black"): "grey";
            }
        }
    }
    Component.onCompleted: {
        pressed.connect(priv.down)
        released.connect(priv.up)
    }

    // private stuff
    QtObject {
        id: priv
        function down(d) {
            cal.selectedDate = d
            if (dragging) {
                updateDrag(d);
                return;
            }
            dragging = true;
            anchor = d;
            updateDrag(d);
        }
        function up(d) {
            updateDrag(d);
            dragging = false;
        }
        function updateDrag(d) {
            var diffDays = function(d1,d2) {
                return Math.round(Math.abs((d1.getTime() - d2.getTime())/(24*60*60*1000)));
            }
            var addDays = function(dt,days)
            {
                var dat = new Date(dt.valueOf());
                dat.setDate(dat.getDate() + days);
                return dat;
            }

            if( d >= anchor) {
                cal.dateStart = anchor;
                cal.dateEnd = d;
            } else {
                cal.dateStart = d;
                cal.dateEnd = anchor;
            }

            if( cal.maxDays >0 ) {
                if( diffDays(cal.dateEnd,cal.dateStart) >= cal.maxDays ) {
                    cal.dateEnd = addDays(cal.dateStart, cal.maxDays-1);
                }
            }

        }

        property bool dragging: false
        property date anchor
    }
}

