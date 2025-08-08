import QtQuick
import QtQuick.Layouts
import QtQuick.Controls
import NetTest

Item {
    DnsQuery {
        id: dnsQuery
    }

    // TextField {
    //     id: domainTextField
    //     Layout.fillWidth: true
    //     placeholderText: qsTr("Domain")
    //     text: "www.google.com"
    // }
    Button {
        text: qsTr("Query DNS")
        anchors.centerIn: parent
        onClicked: {
            dnsQuery.server = "223.5.5.5:53"
            dnsQuery.domain = "dns.alidns.com"
            dnsQuery.type = "A"
            dnsQuery.classType = "IN"
            dnsQuery.startQuery();
        }
    }
}
