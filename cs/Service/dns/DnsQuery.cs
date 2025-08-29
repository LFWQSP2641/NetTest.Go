using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Service.dns;

internal class DnsQuery
{
    private readonly DnsQueryTask _dns = new();

    public Task<string?> DnsQueryAsync(string dnsScheme, string dnsServer, string domain, string recordType = "A", string recordClass = "IN", string sni = "", string clientSubnet = "", string? proxy = null)
    {
        return _dns.DnsQueryAsync(NormalizeDnsServer(dnsScheme, dnsServer), domain, recordType, recordClass, sni, clientSubnet, proxy);
    }

    private string NormalizeDnsServer(string dnsScheme, string dnsServer)
    {
        if (string.IsNullOrWhiteSpace(dnsScheme) || string.IsNullOrWhiteSpace(dnsServer))
            throw new ArgumentException("DNS scheme or server cannot be null or empty.");

        var schemeLower = dnsScheme.ToLowerInvariant();

        // 去除服务器字符串中自带的 scheme（支持任意 like "://"，而不只 Known 列表）
        var schemeSep = dnsServer.IndexOf("://", StringComparison.Ordinal);
        if (schemeSep > 0)
        {
            dnsServer = dnsServer[(schemeSep + 3)..];
        }
        else
        {
            foreach (var known in Global.DnsSchemes)
            {
                if (dnsServer.StartsWith(known, StringComparison.OrdinalIgnoreCase))
                {
                    dnsServer = dnsServer.Substring(known.Length);
                    break;
                }
            }
        }

        // 拆分 hostPort 与 path（path 保留前导 /）
        string hostPort;
        string path;
        var slash = dnsServer.IndexOf('/');
        if (slash >= 0)
        {
            hostPort = dnsServer[..slash];
            path = dnsServer[slash..];
        }
        else
        {
            hostPort = dnsServer;
            path = string.Empty;
        }

        // 根据 scheme 做规范化
        if (IsTlsLike(schemeLower))
        {
            hostPort = EnsurePort(hostPort, 853);
        }
        else if (IsHttpsLike(schemeLower))
        {
            // HTTPS/HTTP3：不改端口，但确保 path
            if (string.IsNullOrEmpty(path))
                path = "/dns-query";
        }
        else
        {
            // 其他（如 udp/tcp）：补 53 端口
            hostPort = EnsurePort(hostPort, 53);
        }

        return dnsScheme + hostPort + path;
    }

    private static bool IsTlsLike(string schemeLower)
        => schemeLower.StartsWith("tls://")
           || schemeLower.StartsWith("quic://")
           || schemeLower.StartsWith("doq://");

    private static bool IsHttpsLike(string schemeLower)
        => schemeLower.StartsWith("https://")
           || schemeLower.StartsWith("https3://")
           || schemeLower.StartsWith("http3://")
           || schemeLower.StartsWith("h3://");

    private static string EnsurePort(string hostPort, int defaultPort)
    {
        if (string.IsNullOrEmpty(hostPort)) return hostPort;

        // IPv6 带方括号形式：[::1] 或 [::1]:853
        if (hostPort.StartsWith("["))
        {
            var end = hostPort.IndexOf(']');
            if (end > 0)
            {
                var after = end + 1;
                if (after < hostPort.Length && hostPort[after] == ':')
                {
                    // 已带端口，校验
                    var portPart = hostPort[(after + 1)..];
                    if (IsValidPort(portPart)) return hostPort;
                    // 端口非法：替换为默认端口
                    return hostPort[..(after + 1)] + defaultPort.ToString();
                }
                // 无端口，追加
                return hostPort + ":" + defaultPort;
            }
            // 不规范的方括号，退化处理：不追加
            return hostPort;
        }

        // 非方括号：可能是 IPv4/域名，或未加括号的 IPv6
        var colonCount = hostPort.Count(c => c == ':');
        if (colonCount == 0)
        {
            // 无端口
            return hostPort + ":" + defaultPort;
        }
        else if (colonCount == 1)
        {
            // 形如 host:port，校验端口
            var idx = hostPort.LastIndexOf(':');
            if (idx >= 0 && idx < hostPort.Length - 1)
            {
                var portPart = hostPort[(idx + 1)..];
                if (IsValidPort(portPart)) return hostPort;
                // 非法端口：替换
                return hostPort[..(idx + 1)] + defaultPort.ToString();
            }
            // 结尾是冒号，直接补默认端口
            return hostPort + defaultPort.ToString();
        }
        else
        {
            // 多个冒号：大概率是未加括号的 IPv6 字面量
            // 若需要添加端口，需加括号
            return $"[{hostPort}]:{defaultPort}";
        }
    }

    private static bool IsValidPort(string s)
        => int.TryParse(s, out var p) && p >= 1 && p <= 65535;
}
