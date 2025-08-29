/// 纯数据实体（Model）：不包含业务逻辑，仅承载字段。
class DnsQueryEntity {
  final String dnsScheme; // 例如 udp:// tcp:// tls:// https:// quic:// https3://
  final String dnsServer;
  final String domain;
  final String recordType;
  final String recordClass;
  final String sni;
  final String clientSubnet;
  final String? proxy;

  const DnsQueryEntity({
    this.dnsScheme = 'udp://',
    this.dnsServer = '223.5.5.5',
    this.domain = 'www.baidu.com',
    this.recordType = 'A',
    this.recordClass = 'IN',
    this.sni = '',
    this.clientSubnet = '',
    this.proxy,
  });

  DnsQueryEntity copyWith({
    String? dnsScheme,
    String? dnsServer,
    String? domain,
    String? recordType,
    String? recordClass,
    String? sni,
    String? clientSubnet,
    String? proxy,
  }) => DnsQueryEntity(
        dnsScheme: dnsScheme ?? this.dnsScheme,
        dnsServer: dnsServer ?? this.dnsServer,
        domain: domain ?? this.domain,
        recordType: recordType ?? this.recordType,
        recordClass: recordClass ?? this.recordClass,
        sni: sni ?? this.sni,
        clientSubnet: clientSubnet ?? this.clientSubnet,
        proxy: proxy ?? this.proxy,
      );
}
