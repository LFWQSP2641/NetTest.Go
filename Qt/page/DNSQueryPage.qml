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
            resultTextArea.log(qsTr("Query failed:") + "\n" + error);
        }
        onQueryFinished: function(hostname, result) {
            resultTextArea.log(qsTr("Query finished:") + "\n" + result);
        }
    }

    ColumnLayout {
        anchors.fill: parent
        spacing: 10

        RowLayout {
            Layout.fillWidth: true
            ComboBox {
                id: netTypeComboBox
                Layout.minimumWidth: netTypeComboBoxTextMetrics.width + implicitIndicatorWidth + leftPadding + rightPadding
                model: ["udp://", "tcp://", "tls://", "https://", "quic://", "https3://"]
                currentIndex: 0
                TextMetrics {
                    id: netTypeComboBoxTextMetrics
                    font: netTypeComboBox.font
                    text: netTypeComboBox.currentText
                }
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
            placeholderText: qsTr("Query Domain")
            text: "www.google.com"
        }

        Text {
            Layout.fillWidth: true
            text: advancedOptions.visible ? qsTr("Advanced Options ▼") : qsTr("Advanced Options ▶")

            MouseArea {
                anchors.fill: parent
                cursorShape: Qt.PointingHandCursor
                onClicked: {
                    advancedOptions.visible = !advancedOptions.visible;
                }
            }
        }

        ColumnLayout {
            id: advancedOptions
            Layout.fillWidth: true
            visible: false
            spacing: 10
            RowLayout {
                ComboBox {
                    id: typeComboBox
                    Layout.minimumWidth: typeComboBoxTextMetrics.width + implicitIndicatorWidth + leftPadding + rightPadding
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
                    TextMetrics {
                        id: typeComboBoxTextMetrics
                        font: typeComboBox.font
                        text: typeComboBox.currentText
                    }
                }

                ComboBox {
                    id: classComboBox
                    Layout.minimumWidth: classComboBoxTextMetrics.width + implicitIndicatorWidth + leftPadding + rightPadding
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
                    TextMetrics {
                        id: classComboBoxTextMetrics
                        font: classComboBox.font
                        text: classComboBox.currentText
                    }
                }
            }

            TextField {
                id: proxyTextField
                Layout.fillWidth: true
                placeholderText: qsTr("socks5 Proxy (optional)")
                text: ""
            }

            TextField {
                id: sniTextField
                Layout.fillWidth: true
                placeholderText: qsTr("TLS SNI (optional)")
                text: ""
            }

            TextField {
                id: clientSubnetTextField
                Layout.fillWidth: true
                placeholderText: qsTr("EDNS Client Subnet (optional)")
                text: ""
            }
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
                dnsQuery.sni = sniTextField.text;
                dnsQuery.clientSubnet = clientSubnetTextField.text;
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
                placeholderText: qsTr("received message and debug log")
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
