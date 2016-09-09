import QtQuick 2.3

Rectangle {
    id: splitamajig
    width: 8
    state: styleData.pressed ? "pressed" : (styleData.hovered ? "hovered":"normal");
    states: [
        State {
            name: "normal"
            PropertyChanges {
                target: splitamajig
                color: "transparent"
                border.width: 0
            }
        },
        State {
            name: "hovered"
            PropertyChanges {
                target: splitamajig
                color: "white"
                border.width: 1
                border.color: "black"
            }
        },
        State {
            name: "pressed"
            PropertyChanges {
                target: splitamajig
                color: "grey"
                border.width: 1
                border.color: "black"
            }
        }

    ]

}

