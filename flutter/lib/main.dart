import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:nettest/ViewModel/dns_query_model.dart';
import 'package:nettest/View/dns_query_page.dart';

void main() {
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  // This widget is the root of your application.
  @override
  Widget build(BuildContext context) {
    return ChangeNotifierProvider(
      create: (_) => DnsQueryModel(),
      child: MaterialApp(
        title: 'NetTest DNS',
        theme: ThemeData(
          colorScheme: ColorScheme.fromSeed(seedColor: Colors.blue),
        ),
        home: DnsQueryPage(),
      ),
    );
  }
}
