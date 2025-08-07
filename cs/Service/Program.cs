using System.Runtime.InteropServices;
using System;

public class Program
{
    [DllImport("netcore.dll", EntryPoint = "DnsRequest", CharSet = CharSet.Ansi)]
    public static extern IntPtr DnsRequest(string dnsServer, string domain, string recordType, string recordClass);

    [DllImport("netcore.dll", EntryPoint = "FreeCString", CharSet = CharSet.Ansi)]
    public static extern void FreeCString(IntPtr ptr);

    public static void Main(string[] args)
    {
        IntPtr result = DnsRequest("223.5.5.5:53", "dns.alidns.com", "A", "IN");
        if (result != IntPtr.Zero)
        {
            string? resultStr = Marshal.PtrToStringUTF8(result);
            if (resultStr != null)
            {
                Console.WriteLine("Result length: " + resultStr.Length);
                Console.WriteLine("Result: " + resultStr);
            }
            else
            {
                Console.WriteLine("Result string is null.");
            }
            FreeCString(result);
        }
        else
        {
            Console.WriteLine("No result returned.");
        }
    }
}