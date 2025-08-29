import 'package:flutter/foundation.dart';

import 'package:nettest/Model/dns_query_entity.dart';
import 'package:nettest/dns/dns_repository.dart';

/// ViewModel（状态与业务流程）：
/// - 暴露可绑定的字段/状态
/// - 组合 Repository 发起请求
/// - 管理 loading/error/result
class DnsQueryModel extends ChangeNotifier {
  DnsQueryModel({DnsRepository? repository, DnsQueryEntity? initial})
      : _repo = repository ?? DnsRepository(),
        _entity = initial ?? const DnsQueryEntity();

  final DnsRepository _repo;
  DnsQueryEntity _entity;

  // 可绑定的 UI 状态
  bool _loading = false;
  String? _error;
  String? _result;

  // 读字段
  String get dnsServer => _entity.dnsServer;
  String get dnsScheme => _entity.dnsScheme;
  String get domain => _entity.domain;
  String get recordType => _entity.recordType;
  String get recordClass => _entity.recordClass;
  String get sni => _entity.sni;
  String get clientSubnet => _entity.clientSubnet;
  String? get proxy => _entity.proxy;

  bool get loading => _loading;
  String? get error => _error;
  String? get result => _result;

  // 写字段（统一通过 copyWith 更新实体）
  set dnsServer(String v) { _entity = _entity.copyWith(dnsServer: v); notifyListeners(); }
  set dnsScheme(String v) { _entity = _entity.copyWith(dnsScheme: v); notifyListeners(); }
  set domain(String v) { _entity = _entity.copyWith(domain: v); notifyListeners(); }
  set recordType(String v) { _entity = _entity.copyWith(recordType: v); notifyListeners(); }
  set recordClass(String v) { _entity = _entity.copyWith(recordClass: v); notifyListeners(); }
  set sni(String v) { _entity = _entity.copyWith(sni: v); notifyListeners(); }
  set clientSubnet(String v) { _entity = _entity.copyWith(clientSubnet: v); notifyListeners(); }
  set proxy(String? v) { _entity = _entity.copyWith(proxy: v); notifyListeners(); }

  Future<void> query() async {
    _loading = true; _error = null; notifyListeners();
    try {
      final r = await _repo.query(
        dnsScheme: _entity.dnsScheme,
        dnsServer: _entity.dnsServer,
        domain: _entity.domain,
        recordType: _entity.recordType,
        recordClass: _entity.recordClass,
        sni: _entity.sni,
        clientSubnet: _entity.clientSubnet,
        proxy: _entity.proxy,
      );
      if (r != null && r.isNotEmpty) {
        _result = (_result == null || _result!.isEmpty)
            ? r
            : (_result! + "\n" + r);
      }
    } catch (e) {
      _error = e.toString();
    } finally {
      _loading = false; notifyListeners();
    }
  }
}
