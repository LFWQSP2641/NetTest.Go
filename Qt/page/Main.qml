import QtQuick

Window {
    width: 640
    height: 480
    visible: true
    title: qsTr("DNS Query Tool")

    DNSQueryPage {
        anchors.fill: parent
        anchors.margins: 10
    }
}
