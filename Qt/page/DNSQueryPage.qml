import QtQuick
import QtQuick.Layouts
import QtQuick.Controls
import NetTest

Item {
    DnsQuery {
        id: dnsQuery
        onBusyChanged: {
            queryButton.enabled = !dnsQuery.busy;
        }
        onQueryFailed: function(hostname, error) {
            resultTextArea.log("Query failed: " + error);
        }
        onQueryFinished: function(hostname, result) {
            resultTextArea.log("Query finished: " + result);
        }
    }

    ColumnLayout {
        anchors.fill: parent

        RowLayout {
            Layout.fillWidth: true
            ComboBox {
                id: netTypeComboBox
                Layout.fillWidth: true
                model: ["udp://", "tcp://", "tls://"]
                currentIndex: 0
            }
            TextField {
                id: serverTextField
                Layout.fillWidth: true
                placeholderText: qsTr("Server")
                text: "223.5.5.5"
            }
        }

        TextField {
            id: domainTextField
            Layout.fillWidth: true
            placeholderText: qsTr("Domain")
            text: "www.google.com"
        }

        RowLayout {
            Layout.fillWidth: true

            ComboBox {
                id: typeComboBox
                Layout.fillWidth: true
                textRole: "text"
                valueRole: "enumValue"
                model: ListModel {
                    ListElement { text: "A"; enumValue: "A" }
                    ListElement { text: "AAAA"; enumValue: "AAAA" }
                    ListElement { text: "NS"; enumValue: "NS" }
                    ListElement { text: "CNAME"; enumValue: "CNAME" }
                    ListElement { text: "SOA"; enumValue: "SOA" }
                    ListElement { text: "PTR"; enumValue: "PTR" }
                    ListElement { text: "MX"; enumValue: "MX" }
                    ListElement { text: "TXT"; enumValue: "TXT" }
                    ListElement { text: "SPF"; enumValue: "SPF" }
                }
                currentIndex: 0
            }

            ComboBox {
                id: classComboBox
                Layout.fillWidth: true
                textRole: "text"
                valueRole: "enumValue"
                model: ListModel {
                    ListElement { text: "IN"; enumValue: "IN" }
                    ListElement { text: "CS"; enumValue: "CS" }
                    ListElement { text: "CH"; enumValue: "CH" }
                    ListElement { text: "HS"; enumValue: "HS" }
                    ListElement { text: "NONE"; enumValue: "NONE" }
                    ListElement { text: "ANY"; enumValue: "ANY" }
                }
                currentIndex: 0
            }
        }

        TextField {
            id: proxyTextField
            Layout.fillWidth: true
            placeholderText: qsTr("socks5 Proxy (optional)")
            text: ""
        }

        Button {
            id: queryButton
            Layout.fillWidth: true
            text: qsTr("Query DNS")
            onClicked: {
                queryButton.enabled = false;
                dnsQuery.server = netTypeComboBox.currentText + serverTextField.text;
                dnsQuery.domain = domainTextField.text;
                dnsQuery.type = typeComboBox.currentValue;
                dnsQuery.classType = classComboBox.currentValue;
                dnsQuery.socks5Server = proxyTextField.text;
                dnsQuery.startQuery();
            }
        }


        Flickable {
            id: flickable
            Layout.fillHeight: true
            Layout.fillWidth: true
            clip: true
            contentHeight: resultTextArea.height
            TextArea {
                id: resultTextArea
                width: parent.width
                placeholderText: "received message and debug log"
                readOnly: true
                wrapMode: TextEdit.Wrap

                function clear() {
                    resultTextArea.text = ""
                }

                function log(text) {
                    resultTextArea.output(Qt.formatDateTime(new Date(), "hh:mm:ss.zzz") + " " + text)
                }

                function output(text) {
                    resultTextArea.append(text)
                    let contentY = flickable.contentHeight - flickable.height
                    flickable.contentY = contentY > 0 ? contentY : 0
                }
            }
        }
    }
}
