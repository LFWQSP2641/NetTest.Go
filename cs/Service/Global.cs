using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Service;

public class Global
{
    public static readonly List<string> DnsSchemes = [
        "udp://", 
        "tcp://", 
        "tls://", 
        "https://",
        "quic://",
        "https3://",
    ];

    public static readonly List<string> DnsRecordType = [
        "A", "AAAA", "CNAME", "MX", "NS", "PTR", "SOA", "SRV", "TXT", "ANY",
        "CAA", "DS", "DNSKEY", "RRSIG", "NSEC", "NSEC3", "NSEC3PARAM",
        "TLSA", "SMIMEA", "SSHFP", "SVCB", "HTTPS"
    ];

    public static readonly List<string> DnsRecordClass = [
        "IN", "CH", "HS", "ANY", "NONE"
    ];

    public static readonly List<string> CommonDnsServers = [
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
    ];
}
