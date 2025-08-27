import QtQuick

Window {
    width: 480
    height: 640
    visible: true
    title: qsTr("DNS Query Tool")

    DNSQueryPage {
        anchors.fill: parent
        anchors.margins: 10
    }
}
