import 'package:nettest/dns/dns_query_task.dart';

/// Repository 层：封装对 FFI 的具体调用，未来可替换为 mock/网络/缓存。
class DnsRepository {
  DnsRepository({DnsQueryTask? task}) : _task = task ?? const DnsQueryTask();

  final DnsQueryTask _task;

  Future<String?> query({
    required String dnsScheme,
    required String dnsServer,
    required String domain,
    String recordType = 'A',
    String recordClass = 'IN',
    String sni = '',
    String clientSubnet = '',
    String? proxy,
  }) {
    final server = _normalizeDnsServer(dnsScheme, dnsServer);
    return _task.query(
      server,
      domain,
      recordType: recordType,
      recordClass: recordClass,
      sni: sni,
      clientSubnet: clientSubnet,
      proxy: proxy,
    );
  }

  String? querySync({
    required String dnsScheme,
    required String dnsServer,
    required String domain,
    String recordType = 'A',
    String recordClass = 'IN',
    String sni = '',
    String clientSubnet = '',
    String? proxy,
  }) {
    final server = _normalizeDnsServer(dnsScheme, dnsServer);
    return _task.querySync(
      server,
      domain,
      recordType: recordType,
      recordClass: recordClass,
      sni: sni,
      clientSubnet: clientSubnet,
      proxy: proxy,
    );
  }

  // Mirror C# NormalizeDnsServer behavior.
  static String _normalizeDnsServer(String dnsScheme, String dnsServer) {
    if (dnsScheme.trim().isEmpty || dnsServer.trim().isEmpty) {
      throw ArgumentError('DNS scheme or server cannot be null or empty.');
    }

    final schemeLower = dnsScheme.toLowerCase();

    // Strip any scheme already present in server
    final idx = dnsServer.indexOf('://');
    if (idx > 0) {
      dnsServer = dnsServer.substring(idx + 3);
    } else {
      for (final known in _dnsSchemes) {
        if (dnsServer.toLowerCase().startsWith(known)) {
          dnsServer = dnsServer.substring(known.length);
          break;
        }
      }
    }

    // Split hostPort and path
    String hostPort;
    String path;
    final slash = dnsServer.indexOf('/');
    if (slash >= 0) {
      hostPort = dnsServer.substring(0, slash);
      path = dnsServer.substring(slash);
    } else {
      hostPort = dnsServer;
      path = '';
    }

    if (_isTlsLike(schemeLower)) {
      hostPort = _ensurePort(hostPort, 853);
    } else if (_isHttpsLike(schemeLower)) {
      if (path.isEmpty) path = '/dns-query';
    } else {
      hostPort = _ensurePort(hostPort, 53);
    }

    return dnsScheme + hostPort + path;
  }

  static bool _isTlsLike(String schemeLower) =>
      schemeLower.startsWith('tls://') ||
      schemeLower.startsWith('quic://') ||
      schemeLower.startsWith('doq://');

  static bool _isHttpsLike(String schemeLower) =>
      schemeLower.startsWith('https://') ||
      schemeLower.startsWith('https3://') ||
      schemeLower.startsWith('http3://') ||
      schemeLower.startsWith('h3://');

  static String _ensurePort(String hostPort, int defaultPort) {
    if (hostPort.isEmpty) return hostPort;

    if (hostPort.startsWith('[')) {
      final end = hostPort.indexOf(']');
      if (end > 0) {
        final after = end + 1;
        if (after < hostPort.length && hostPort[after] == ':') {
          final portPart = hostPort.substring(after + 1);
          if (_isValidPort(portPart)) return hostPort;
          return hostPort.substring(0, after + 1) + defaultPort.toString();
        }
  return '$hostPort:$defaultPort';
      }
      return hostPort;
    }

    final colonCount = hostPort.split('').where((c) => c == ':').length;
    if (colonCount == 0) {
      return '$hostPort:$defaultPort';
    } else if (colonCount == 1) {
      final idx = hostPort.lastIndexOf(':');
      if (idx >= 0 && idx < hostPort.length - 1) {
        final portPart = hostPort.substring(idx + 1);
        if (_isValidPort(portPart)) return hostPort;
        return '${hostPort.substring(0, idx + 1)}$defaultPort';
      }
  return '$hostPort$defaultPort';
    } else {
  return '[$hostPort]:$defaultPort';
    }
  }

  static bool _isValidPort(String s) {
    final p = int.tryParse(s);
    return p != null && p >= 1 && p <= 65535;
  }

  static const List<String> _dnsSchemes = <String>[
    'udp://', 'tcp://', 'tls://', 'https://', 'quic://', 'https3://',
  ];
}
