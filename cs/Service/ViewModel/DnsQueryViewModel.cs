using ReactiveUI;
using ReactiveUI.SourceGenerators;
using Service.dns;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Reactive;
using System.Reactive.Disposables;
using System.Reactive.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Threading;
using System.Collections.ObjectModel;

namespace Service.ViewModel;

public partial class DnsQueryViewModel : ReactiveObject, IDisposable
{
    [Reactive]
    private string _dnsServerScheme = Global.DnsSchemes.First();

    [Reactive]
    private string _dnsServer = Global.CommonDnsServers.First();

    [Reactive]
    private string _domain = "www.baidu.com";

    [Reactive]
    private string _recordType = Global.DnsRecordType.First();

    [Reactive]
    private string _recordClass = Global.DnsRecordClass.First();

    [Reactive]
    private string? _sni;

    [Reactive]
    private string? _clientSubnet;

    [Reactive]
    private string? _proxy;

    [ObservableAsProperty]
    private bool _isBusy;

    [ObservableAsProperty]
    private bool _canQuery;

    [ObservableAsProperty]
    private string? _result;

    [Reactive]
    private string? _error;

    // 追加显示的结果日志（用于不覆盖而是累加）
    [Reactive]
    private string? _resultLog;

    // 按项展示的结果集合
    public ObservableCollection<LogEntry> Entries { get; } = new();

    public sealed class LogEntry
    {
        public required string Time { get; init; }
        public required string Title { get; init; }
        public required string Body { get; init; }
        public bool IsError { get; init; }
        public string Header => Time + " " + Title;
    }

    public ReactiveCommand<Unit, string?> QueryCommand { get; }

    private readonly DnsQuery _dns = new();
    private readonly CompositeDisposable _disposables = new();

    public DnsQueryViewModel()
    {
        // 校验/CanExecute：只要关键字段非空即可查询
        var canQuery = this
            .WhenAnyValue(vm => vm.DnsServer, vm => vm.Domain, vm => vm.RecordType,
                (server, domain, type) =>
                    !string.IsNullOrWhiteSpace(server)
                    && !string.IsNullOrWhiteSpace(domain)
                    && !string.IsNullOrWhiteSpace(type))
            .DistinctUntilChanged();

        _canQueryHelper = canQuery
            .ObserveOn(RxApp.MainThreadScheduler)
            .ToProperty(this, vm => vm.CanQuery)
            .DisposeWith(_disposables);

        // 异步命令：集中并发、忙碌状态与异常处理
        QueryCommand = ReactiveCommand.CreateFromTask<string?>(ExecuteQueryAsync, canQuery);
        _isBusyHelper = QueryCommand.IsExecuting
            .ObserveOn(RxApp.MainThreadScheduler)
            .ToProperty(this, vm => vm.IsBusy)
            .DisposeWith(_disposables);

        // 将命令输出投递为只读属性 Result
        _resultHelper = QueryCommand
            .ObserveOn(RxApp.MainThreadScheduler)
            .ToProperty(this, vm => vm.Result)
            .DisposeWith(_disposables);

        // 将每次查询结果追加到 ResultLog
        QueryCommand
            .Where(r => r is not null)
            .ObserveOn(RxApp.MainThreadScheduler)
            .Subscribe(r =>
            {
                var text = r!;
                ResultLog = string.IsNullOrEmpty(ResultLog)
                    ? text
                    : ResultLog + Environment.NewLine + text;

                Entries.Add(new LogEntry
                {
                    Time = DateTime.Now.ToString("HH:mm:ss.fff"),
                    Title = "Query finished:",
                    Body = text,
                    IsError = false
                });
            })
            .DisposeWith(_disposables);

        // 统一处理异常
        QueryCommand.ThrownExceptions
            .ObserveOn(RxApp.MainThreadScheduler)
            .Subscribe(ex =>
            {
                Error = ex.Message;
                Entries.Add(new LogEntry
                {
                    Time = DateTime.Now.ToString("HH:mm:ss.fff"),
                    Title = "Query failed:",
                    Body = ex.Message,
                    IsError = true
                });
            })
            .DisposeWith(_disposables);

    }

    // ReactiveCommand 支持 CancellationToken（由框架注入）
    private async Task<string?> ExecuteQueryAsync(CancellationToken ct)
    {
        Error = null; // 清空上一条错误
        // 目前 DnsQueryTask 不支持取消；这里的 ct 仅用于早退，避免结果写回
        if (ct.IsCancellationRequested) return null;

        var result = await _dns.DnsQueryAsync(
            dnsScheme: DnsServerScheme,
            dnsServer: DnsServer,
            domain: Domain,
            recordType: string.IsNullOrWhiteSpace(RecordType) ? "A" : RecordType,
            recordClass: string.IsNullOrWhiteSpace(RecordClass) ? "IN" : RecordClass,
            sni: Sni ?? "",
            clientSubnet: ClientSubnet ?? "",
            proxy: string.IsNullOrWhiteSpace(Proxy) ? null : Proxy
        ).ConfigureAwait(false);

        if (ct.IsCancellationRequested) return null;
        return result;
    }

    public void Dispose() => _disposables.Dispose();
}
