import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:nettest/ViewModel/dns_query_model.dart';

class DnsQueryPage extends StatelessWidget {
  const DnsQueryPage({super.key});
  @override
  Widget build(BuildContext context) {
    final vm = context.watch<DnsQueryModel>();
    return Scaffold(
      appBar: AppBar(title: const Text('DNS Query')),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          children: [
            Row(
              children: [
                const SizedBox(width: 110, child: Text('Scheme')),
                Expanded(
                  child: DropdownMenu<String>(
                    initialSelection: vm.dnsScheme,
                    dropdownMenuEntries: const [
                      DropdownMenuEntry(value: 'udp://', label: 'udp://'),
                      DropdownMenuEntry(value: 'tcp://', label: 'tcp://'),
                      DropdownMenuEntry(value: 'tls://', label: 'tls://'),
                      DropdownMenuEntry(value: 'https://', label: 'https://'),
                      DropdownMenuEntry(value: 'quic://', label: 'quic://'),
                      DropdownMenuEntry(value: 'https3://', label: 'https3://'),
                    ],
                    onSelected: (v) {
                      if (v != null) vm.dnsScheme = v;
                    },
                  ),
                ),
              ],
            ),
            _RowField(
              label: 'Server',
              initial: vm.dnsServer,
              onChanged: (v) => vm.dnsServer = v,
            ),
            _RowField(
              label: 'Domain',
              initial: vm.domain,
              onChanged: (v) => vm.domain = v,
            ),
            Row(
              children: [
                Expanded(
                  child: _RowField(
                    label: 'Type',
                    initial: vm.recordType,
                    onChanged: (v) => vm.recordType = v,
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: _RowField(
                    label: 'Class',
                    initial: vm.recordClass,
                    onChanged: (v) => vm.recordClass = v,
                  ),
                ),
              ],
            ),
            _RowField(
              label: 'SNI',
              initial: vm.sni,
              onChanged: (v) => vm.sni = v,
            ),
            _RowField(
              label: 'Client Subnet',
              initial: vm.clientSubnet,
              onChanged: (v) => vm.clientSubnet = v,
            ),
            _RowField(
              label: 'SOCKS5',
              hint: 'socks5://host:port',
              initial: vm.proxy ?? '',
              onChanged: (v) => vm.proxy = v.isEmpty ? null : v,
            ),
            const SizedBox(height: 12),
            Row(
              children: [
                ElevatedButton.icon(
                  onPressed: vm.loading ? null : () => vm.query(),
                  icon: const Icon(Icons.search),
                  label: const Text('Query'),
                ),
                const SizedBox(width: 12),
                if (vm.loading) const SizedBox.square(dimension: 20, child: CircularProgressIndicator(strokeWidth: 2)),
              ],
            ),
            const Divider(height: 24),
            Expanded(
              child: Container(
                width: double.infinity,
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  border: Border.all(color: Theme.of(context).dividerColor),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: vm.error != null
                    ? Text(vm.error!, style: const TextStyle(color: Colors.red))
                    : SelectableText(vm.result ?? 'No result'),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _RowField extends StatelessWidget {
  const _RowField({
    required this.label,
    required this.initial,
    required this.onChanged,
    this.hint,
  });
  final String label;
  final String initial;
  final String? hint;
  final ValueChanged<String> onChanged;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Row(
        children: [
          SizedBox(width: 110, child: Text(label)),
          Expanded(
            child: TextFormField(
              initialValue: initial,
              decoration: InputDecoration(hintText: hint),
              onChanged: onChanged,
            ),
          ),
        ],
      ),
    );
  }
}