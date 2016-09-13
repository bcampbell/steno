import QtQml 2.0
import QtQuick 2.3
import QtQuick.Controls 1.2
import QtQuick.Layouts 1.0

import "helper.js" as Helper

// ArtInfo shows an individual article

Item {
    id: artInfo
    // current article (null=none)
    property var art: null

    // array of strings to highlight in the text
    property var hlTerms: []

    property string hoveredLink: {
        if (content.hoveredLink) { return content.hoveredLink; }
        if (canonicalURL.hoveredLink) { return canonicalURL.hoveredLink; }
        if (urls.hoveredLink) { return urls.hoveredLink; }
        return "";
    }

    function showArt(a,highlightTerms) {

        // causes multiple refreshes... not sure if there's a way to
        // bundle signaled changes, but not a big deal.
        hlTerms = highlightTerms;
        art = a; 
    }

    ColumnLayout {
        anchors.fill: parent
        visible: (art != null )

        Text {
            Layout.fillWidth: true
            Layout.margins: 8
            text: art ? art.headline : ""
            font.bold: true
            font.pixelSize: 24
            elide: Text.ElideRight
            maximumLineCount: 2
            wrapMode: Text.Wrap
        }

        RowLayout {

            // article content
            TextArea {
                //Layout.minimumWidth: 40
                Layout.margins: 8
                id: content 
                Layout.fillWidth: true
                Layout.fillHeight: true
                Layout.preferredWidth: 800
                text: art ? ctrl.renderContent(art,hlTerms) : "";
                readOnly: true

                onLinkActivated: app.browseURL(link)

                wrapMode: Text.WordWrap
                textFormat: Text.RichText

                menu: Menu {
                    id: editMenu
                    title: "Edit"

                    MenuItem {
                        text: "Copy"
                        shortcut: "Ctrl+C"
                        enabled: content.selectedText!=""
                        onTriggered: { content.copy(); }
                    }
                }
            }

            // article metadata
            ColumnLayout {
                Layout.alignment: Qt.AlignTop
                Layout.minimumWidth: 40
                Layout.margins: 8
                Layout.fillHeight: true;
                Layout.fillWidth: true;

                Label { text: "CanonicalURL:" }
                Text{
                    id: canonicalURL
                    Layout.leftMargin: 16
                    text: art ? Helper.markupLinks(art.canonicalURL) : ""
                    onLinkActivated: app.browseURL(link)
                    MouseArea {
                        acceptedButtons: Qt.NoButton
                        cursorShape: parent.hoveredLink=="" ? Qt.ArrorCursor : Qt.PointingHandCursor
                        anchors.fill: parent
                    }
                }

                Label { text: "URLs:" }
                Repeater {
                    id: urls 
                    property string hoveredLink: ""
                    model: art ? art.numURLs() : 0
                    Text{
                        Layout.leftMargin: 16
                        text: Helper.markupLinks(art.getURL(index))
                        onLinkActivated: app.browseURL(link)
                        onHoveredLinkChanged: { urls.hoveredLink = hoveredLink; }
/*
                        MouseArea {
                            acceptedButtons: Qt.NoButton
                            cursorShape: parent.hoveredLink=="" ? Qt.ArrorCursor : Qt.PointingHandCursor;
                            anchors.fill: parent;
                        }
*/
                    }
                }

                Label { text: "Byline:" }
                Text{
                    Layout.leftMargin: 16
                    text: art ? art.byline : ""
                }

                Label { text: "Published:" }
                Text{
                    Layout.leftMargin: 16
                    text: art ? art.published : ""
                }

                Label { text: "Updated:" }
                Text{
                    Layout.leftMargin: 16
                    text: art ? art.updated : ""
                }

                Label { text: "Published:" }
                Text{
                    Layout.leftMargin: 16
                    text: art ? art.published : ""
                }

                Label { text: "Section:" }
                Text{
                    Layout.leftMargin: 16
                    text: art ? art.section : ""
                }

                Label { text: "Pub:" }
                Text{
                    Layout.leftMargin: 16
                    text: art ? art.pub : ""
                }

                Label { text: "Keywords:" }
                Text{
                    Layout.leftMargin: 16
                    text: art ? art.keywordsString() : ""
                }

                Label { text: "Tags:" }
                Text{
                    Layout.leftMargin: 16
                    text: art ? art.tagsString() : ""
                }
            }

        }
    }

}

