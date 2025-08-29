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
            ComboBox {
                id: serverComboBox
                Layout.fillWidth: true
                editable: true
                model: [
                    "223.5.5.5",          // 阿里 DNS
                    "119.29.29.29",       // 腾讯 DNS
                    "dns.alidns.com",     // 阿里 DNS 域名
                    "doh.pub",            // 腾讯 DNS 域名
                    "8.8.8.8",            // Google DNS
                    "1.1.1.1",            // Cloudflare DNS
                    "dns.google",         // Google DNS 域名
                    "cloudflare-dns.com", // Cloudflare DNS 域名
                    "180.76.76.76",       // 百度 DNS
                    "114.114.114.114",    // 114 DNS
                    "1.2.4.8",            // CNNIC DNS
                ]
                currentIndex: 0
                inputMethodHints: Qt.ImhUrlCharactersOnly
            }
        }

        TextField {
            id: domainTextField
            Layout.fillWidth: true
            placeholderText: qsTr("Query Domain")
            inputMethodHints: Qt.ImhUrlCharactersOnly
            text: "www.baidu.com"
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
                    editable: true
                    model: ["A", "AAAA", "NS", "CNAME", "SOA", "PTR", "MX", "TXT", "SPF", "SRV", "CAA", "ANY", "DNSKEY", "DS", "RRSIG"]
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
                    editable: true
                    model: ["IN", "CS", "CH", "HS", "NONE", "ANY"]
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
                inputMethodHints: Qt.ImhUrlCharactersOnly
                text: ""
            }

            TextField {
                id: sniTextField
                Layout.fillWidth: true
                placeholderText: qsTr("TLS SNI (optional)")
                inputMethodHints: Qt.ImhUrlCharactersOnly
                text: ""
            }

            TextField {
                id: clientSubnetTextField
                Layout.fillWidth: true
                placeholderText: qsTr("EDNS Client Subnet (optional)")
                inputMethodHints: Qt.ImhUrlCharactersOnly
                text: ""
            }
        }

        Button {
            id: queryButton
            Layout.fillWidth: true
            text: qsTr("Query DNS")
            onClicked: {
                queryButton.enabled = false;
                dnsQuery.server = netTypeComboBox.currentText + serverComboBox.currentText;
                dnsQuery.domain = domainTextField.text;
                dnsQuery.type = typeComboBox.currentText;
                dnsQuery.classType = classComboBox.currentText;
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
