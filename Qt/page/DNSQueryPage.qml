import QtQuick
import QtQuick.Layouts
import QtQuick.Controls
import NetTest
import QtQuick.Controls.Material

Item {
    // 用于按项存储并展示查询结果
    ListModel {
        id: resultsModel
    }

    DnsQuery {
        id: dnsQuery
        onBusyChanged: {
            queryButton.enabled = !dnsQuery.busy;
        }
        onQueryFailed: function(hostname, error) {
            resultsModel.append({
                                   time: Qt.formatDateTime(new Date(), "hh:mm:ss.zzz"),
                                   title: qsTr("Query failed:"),
                                   body: error,
                                   isError: true
                               });
            resultsList.scrollToBottom();
        }
        onQueryFinished: function(hostname, result) {
            resultsModel.append({
                                   time: Qt.formatDateTime(new Date(), "hh:mm:ss.zzz"),
                                   title: qsTr("Query finished:"),
                                   body: result,
                                   isError: false
                               });
            resultsList.scrollToBottom();
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

        Item {
            Layout.fillHeight: true
            Layout.fillWidth: true
            clip: true

            // 占位文本（无结果时）
            Text {
                anchors.centerIn: parent
                visible: resultsModel.count === 0
                text: qsTr("received message and debug log")
                color: "#808080"
                wrapMode: Text.Wrap
                width: parent.width * 0.9
                horizontalAlignment: Text.AlignHCenter
            }

            ListView {
                id: resultsList
                anchors.fill: parent
                model: resultsModel
                clip: true
                spacing: 8

                delegate: Rectangle {
                    width: resultsList.width
                    radius: 4
                    // 交替背景色（条纹效果）
                    color: (index % 2 === 0) ? Material.background.lighter(1.05) : Material.background.darker(1.05)
                    border.width: 1
                    border.color: "transparent"

                    property int pad: 8
                    height: contentCol.implicitHeight + pad * 2

                    Column {
                        id: contentCol
                        anchors.fill: parent
                        anchors.margins: parent.pad
                        spacing: 4

                        Text {
                            text: time + " " + title
                            font.bold: true
                            color: isError ? "red" : Material.accent
                            wrapMode: Text.Wrap
                        }
                        Text {
                            text: body
                            color: isError ? "red" : Material.foreground
                            wrapMode: Text.Wrap
                        }
                    }
                }

                function scrollToBottom() {
                    if (resultsModel.count > 0) {
                        resultsList.positionViewAtEnd();
                    }
                }
            }
        }
    }
}
