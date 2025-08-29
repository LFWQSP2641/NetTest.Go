// Dart FFI wrapper for netcore DNS functions with async callback support.
// Compatible with Windows (netcore.dll), Android (.so), Linux/macOS fallbacks.

// ignore_for_file: constant_identifier_names

import 'dart:async';
import 'dart:ffi' as ffi;
import 'dart:isolate';
import 'dart:io' show Platform;

import 'package:ffi/ffi.dart' as pkf;

typedef _CCharPtr = ffi.Pointer<pkf.Utf8>;
typedef _CVoidPtr = ffi.Pointer<ffi.Void>;

// Native typedefs
typedef _DnsRequestNative = _CCharPtr Function(
	_CCharPtr dnsServer,
	_CCharPtr domain,
	_CCharPtr recordType,
	_CCharPtr recordClass,
	_CCharPtr sni,
	_CCharPtr clientSubnet,
);

typedef _DnsRequestAsyncNative = ffi.Void Function(
	_CCharPtr dnsServer,
	_CCharPtr domain,
	_CCharPtr recordType,
	_CCharPtr recordClass,
	_CCharPtr sni,
	_CCharPtr clientSubnet,
	ffi.Pointer<ffi.NativeFunction<_DnsRequestCallbackNative>> callback,
	_CVoidPtr userData,
);

typedef _DnsRequestOverSocks5Native = _CCharPtr Function(
	_CCharPtr proxy,
	_CCharPtr dnsServer,
	_CCharPtr domain,
	_CCharPtr recordType,
	_CCharPtr recordClass,
	_CCharPtr sni,
	_CCharPtr clientSubnet,
);

typedef _DnsRequestOverSocks5AsyncNative = ffi.Void Function(
	_CCharPtr proxy,
	_CCharPtr dnsServer,
	_CCharPtr domain,
	_CCharPtr recordType,
	_CCharPtr recordClass,
	_CCharPtr sni,
	_CCharPtr clientSubnet,
	ffi.Pointer<ffi.NativeFunction<_DnsRequestCallbackNative>> callback,
	_CVoidPtr userData,
);

typedef _FreeCStringNative = ffi.Void Function(_CCharPtr);

// Dart-side signatures
typedef _DnsRequestDart = _CCharPtr Function(
	_CCharPtr,
	_CCharPtr,
	_CCharPtr,
	_CCharPtr,
	_CCharPtr,
	_CCharPtr,
);
typedef _DnsRequestAsyncDart = void Function(
	_CCharPtr,
	_CCharPtr,
	_CCharPtr,
	_CCharPtr,
	_CCharPtr,
	_CCharPtr,
	ffi.Pointer<ffi.NativeFunction<_DnsRequestCallbackNative>>,
	_CVoidPtr,
);
typedef _DnsRequestOverSocks5Dart = _CCharPtr Function(
	_CCharPtr,
	_CCharPtr,
	_CCharPtr,
	_CCharPtr,
	_CCharPtr,
	_CCharPtr,
	_CCharPtr,
);
typedef _DnsRequestOverSocks5AsyncDart = void Function(
	_CCharPtr,
	_CCharPtr,
	_CCharPtr,
	_CCharPtr,
	_CCharPtr,
	_CCharPtr,
	_CCharPtr,
	ffi.Pointer<ffi.NativeFunction<_DnsRequestCallbackNative>>,
	_CVoidPtr,
);
typedef _FreeCStringDart = void Function(_CCharPtr);

// Callback typedef
typedef _DnsRequestCallbackNative = ffi.Void Function(
	_CVoidPtr userData,
	_CCharPtr result,
);
// Dart-side callback typedef not required (we use fromFunction with native type).

class _NativeBindings {
  _NativeBindings._(ffi.DynamicLibrary lib)
	  : dnsRequest = lib
			.lookupFunction<_DnsRequestNative, _DnsRequestDart>('DnsRequest'),
		dnsRequestAsync = lib.lookupFunction<
						_DnsRequestAsyncNative,
						_DnsRequestAsyncDart>('DnsRequestAsync'),
		dnsRequestOverSocks5 = lib.lookupFunction<
						_DnsRequestOverSocks5Native,
						_DnsRequestOverSocks5Dart>('DnsRequestOverSocks5'),
		dnsRequestOverSocks5Async = lib.lookupFunction<
						_DnsRequestOverSocks5AsyncNative,
						_DnsRequestOverSocks5AsyncDart>('DnsRequestOverSocks5Async'),
		freeCString =
			lib.lookupFunction<_FreeCStringNative, _FreeCStringDart>('FreeCString');
	final _DnsRequestDart dnsRequest;
	final _DnsRequestAsyncDart dnsRequestAsync;
	final _DnsRequestOverSocks5Dart dnsRequestOverSocks5;
	final _DnsRequestOverSocks5AsyncDart dnsRequestOverSocks5Async;
	final _FreeCStringDart freeCString;
}

_NativeBindings _loadBindings() {
	ffi.DynamicLibrary lib;
	if (Platform.isWindows) {
		// Expect netcore.dll to be on PATH or beside the exe.
		lib = ffi.DynamicLibrary.open('netcore.dll');
	} else if (Platform.isAndroid) {
		// The .so must be packaged in app/src/main/jniLibs/<abi>/libnetcore.so
		// and loaded by base name without prefix.
		// If your .so is named differently, adjust here accordingly.
		// Try common names in order.
		ffi.DynamicLibrary? ok;
		for (final name in const [
			'libnetcore.so',
			'netcore',
			'netcore.so',
			// Project-specific artifacts (fallbacks):
			'libandroidnetcore_arm64-v8a.so',
			'libandroidnetcore_armeabi-v7a.so',
			'libandroidnetcore_x86_64.so',
			'libandroidnetcore_x86.so',
		]) {
			try {
				ok = ffi.DynamicLibrary.open(name);
				break;
			} catch (_) {
				// try next
			}
		}
		if (ok == null) {
			throw StateError(
				'Failed to load netcore .so on Android. Ensure libnetcore.so is packaged under jniLibs.',
			);
		}
		lib = ok;
	} else if (Platform.isLinux) {
		lib = ffi.DynamicLibrary.open('libnetcore.so');
	} else if (Platform.isMacOS) {
		lib = ffi.DynamicLibrary.open('libnetcore.dylib');
	} else {
		throw UnsupportedError('Unsupported platform for netcore FFI');
	}
	return _NativeBindings._(lib);
}

final _NativeBindings _bindings = _loadBindings();

class DnsQueryTask {
	const DnsQueryTask();

	// High-level async API mirroring the C# version.
	Future<String?> query(
		String dnsServer,
		String domain, {
		String recordType = 'A',
		String recordClass = 'IN',
		String sni = '',
		String clientSubnet = '',
		String? proxy,
	}) async {
		// 为规避 Dart VM 对原生回调线程的限制，改为在后台 Isolate 调用同步 FFI。
		return Isolate.run(() => querySync(
			dnsServer,
			domain,
			recordType: recordType,
			recordClass: recordClass,
			sni: sni,
			clientSubnet: clientSubnet,
			proxy: proxy,
		));
	}

	// Synchronous variant. Blocks the calling isolate until native returns.
	String? querySync(
		String dnsServer,
		String domain, {
		String recordType = 'A',
		String recordClass = 'IN',
		String sni = '',
		String clientSubnet = '',
		String? proxy,
	}) {
		final pDns = dnsServer.toNativeUtf8();
		final pDom = domain.toNativeUtf8();
		final pType = recordType.toNativeUtf8();
		final pClass = recordClass.toNativeUtf8();
		final pSni = sni.toNativeUtf8();
		final pSubnet = clientSubnet.toNativeUtf8();
		_CCharPtr? pProxy;
		if (proxy != null) pProxy = proxy.toNativeUtf8();

		_CCharPtr resultPtr = ffi.nullptr;
		try {
			if (pProxy == null) {
				resultPtr = _bindings.dnsRequest(
					pDns,
					pDom,
					pType,
					pClass,
					pSni,
					pSubnet,
				);
			} else {
				resultPtr = _bindings.dnsRequestOverSocks5(
					pProxy,
					pDns,
					pDom,
					pType,
					pClass,
					pSni,
					pSubnet,
				);
			}
			if (resultPtr == ffi.nullptr) return null;
			final result = resultPtr.toDartString();
			return result;
		} finally {
			// Free inputs
			pkf.malloc
				..free(pDns)
				..free(pDom)
				..free(pType)
				..free(pClass)
				..free(pSni)
				..free(pSubnet);
			if (pProxy != null) pkf.malloc.free(pProxy);
			// Free native result string
			if (resultPtr != ffi.nullptr) {
				_bindings.freeCString(resultPtr);
			}
		}
	}
}

// No extra extensions: Pointer<Utf8>.toDartString() is provided by package:ffi

