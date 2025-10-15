import QtQuick

Window {
    width: 480
    height: 640
    visible: true
    title: qsTr("DNS Query Tool")

    DNSQueryPage {
        anchors {
            left: parent.left; right: parent.right; top: parent.top
            topMargin: 10 + parent.SafeArea.margins.top
            leftMargin: 10 + parent.SafeArea.margins.left
            rightMargin: 10 + parent.SafeArea.margins.right
        }
    }
}
