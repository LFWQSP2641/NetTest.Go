using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.Linq;
using System.Runtime.InteropServices;
using System.Text;
using System.Threading.Tasks;

namespace Service.dns;

internal class DnsQueryTask
{
    [UnmanagedFunctionPointer(CallingConvention.Cdecl)]
    private delegate void DnsRequestCallback(IntPtr userData, IntPtr result);

    [DllImport("netcore", EntryPoint = "DnsRequest", CharSet = CharSet.Ansi, CallingConvention = CallingConvention.Cdecl)]
    private static extern IntPtr DnsRequest(string dnsServer, string domain, string recordType, string recordClass, string sni, string clientSubnet);

    [DllImport("netcore", EntryPoint = "DnsRequestAsync", CharSet = CharSet.Ansi, CallingConvention = CallingConvention.Cdecl)]
    private static extern void DnsRequestAsync(string dnsServer, string domain, string recordType, string recordClass, string sni, string clientSubnet, DnsRequestCallback callback, IntPtr userData);

    [DllImport("netcore", EntryPoint = "DnsRequestOverSocks5", CharSet = CharSet.Ansi, CallingConvention = CallingConvention.Cdecl)]
    private static extern IntPtr DnsRequestOverSocks5(string proxy, string dnsServer, string domain, string recordType, string recordClass, string sni, string clientSubnet);

    [DllImport("netcore", EntryPoint = "DnsRequestOverSocks5Async", CharSet = CharSet.Ansi, CallingConvention = CallingConvention.Cdecl)]
    private static extern void DnsRequestOverSocks5Async(string proxy, string dnsServer, string domain, string recordType, string recordClass, string sni, string clientSubnet, DnsRequestCallback callback, IntPtr userData);

    [DllImport("netcore", EntryPoint = "FreeCString", CharSet = CharSet.Ansi, CallingConvention = CallingConvention.Cdecl)]
    private static extern void FreeCString(IntPtr ptr);

    // Keep a rooted delegate to avoid GC of the callback while native code holds it.
    private static readonly DnsRequestCallback s_callback = DnsRequestCallbackImpl;

    private sealed class RequestState
    {
        public required DnsQueryTask TaskInstance { get; init; }
        public required TaskCompletionSource<string?> Tcs { get; init; }
    }

    public Task<string?> DnsQueryAsync(string dnsServer, string domain, string recordType = "A", string recordClass = "IN", string sni = "", string clientSubnet = "", string? proxy = null)
    {
        var tcs = new TaskCompletionSource<string?>(TaskCreationOptions.RunContinuationsAsynchronously);
        var state = new RequestState
        {
            TaskInstance = this,
            Tcs = tcs
        };

        GCHandle handle = GCHandle.Alloc(state);
        IntPtr userData = GCHandle.ToIntPtr(handle);
        try
        {
            if (string.IsNullOrEmpty(proxy))
            {
                DnsRequestAsync(dnsServer, domain, recordType, recordClass, sni, clientSubnet, s_callback, userData);
            }
            else
            {
                DnsRequestOverSocks5Async(proxy, dnsServer, domain, recordType, recordClass, sni, clientSubnet, s_callback, userData);
            }
        }
        catch (Exception ex)
        {
            if (handle.IsAllocated)
            {
                handle.Free();
            }
            tcs.TrySetException(ex);
        }

        return tcs.Task;
    }

    public string? DnsQuery(string dnsServer, string domain, string recordType = "A", string recordClass = "IN", string sni = "", string clientSubnet = "", string? proxy = null)
    {
        IntPtr resultPtr;
        if (string.IsNullOrEmpty(proxy))
        {
            resultPtr = DnsRequest(dnsServer, domain, recordType, recordClass, sni, clientSubnet);
        }
        else
        {
            resultPtr = DnsRequestOverSocks5(proxy, dnsServer, domain, recordType, recordClass, sni, clientSubnet);
        }
        if (resultPtr == IntPtr.Zero)
        {
            return null;
        }
        string? resultString = Marshal.PtrToStringAnsi(resultPtr);
        FreeCString(resultPtr);
        return resultString;
    }

    private void HandleDnsResponse(string result)
    {
        Trace.TraceInformation($"DnsQueryTask.HandleDnsResponse: {result}");
    }

    private static void DnsRequestCallbackImpl(IntPtr userData, IntPtr result)
    {
        GCHandle handle = GCHandle.FromIntPtr(userData);
        RequestState? state = null;
        try
        {
            state = handle.Target as RequestState;
            var taskInstance = state?.TaskInstance;
            string? resultString = result == IntPtr.Zero ? null : Marshal.PtrToStringAnsi(result);
            if (taskInstance != null && resultString != null)
            {
                taskInstance.HandleDnsResponse(resultString);
            }
            state?.Tcs.TrySetResult(resultString);
        }
        finally
        {
            if (handle.IsAllocated)
            {
                handle.Free();
            }
            if (result != IntPtr.Zero)
            {
                FreeCString(result);
            }
        }
    }
}
