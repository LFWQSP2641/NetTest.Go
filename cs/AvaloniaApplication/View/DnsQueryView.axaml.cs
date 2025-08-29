using Avalonia.ReactiveUI;
using Avalonia.Threading;
using ReactiveUI;
using Service;
using Service.ViewModel;
using System.Reactive.Disposables;
using System.Reactive.Linq;

namespace AvaloniaApplication.View;

public partial class DnsQueryView : ReactiveUserControl<DnsQueryViewModel>
{
    public DnsQueryView()
    {
        InitializeComponent();

        ViewModel = new DnsQueryViewModel();

        DnsSchemeComboBox.ItemsSource = Global.DnsSchemes;
        DnsServerComboBox.ItemsSource = Global.CommonDnsServers;
        RecordTypeComboBox.ItemsSource = Global.DnsRecordType;
        RecordClassComboBox.ItemsSource = Global.DnsRecordClass;

        DnsServerComboBox.Text = ViewModel.DnsServer;
        RecordTypeComboBox.Text = ViewModel.RecordType;
        RecordClassComboBox.Text = ViewModel.RecordClass;

        this.WhenActivated(disposables =>
        {
            this.Bind(ViewModel, vm => vm.DnsServerScheme, v => v.DnsSchemeComboBox.SelectedValue)
                .DisposeWith(disposables);
            this.Bind(ViewModel, vm => vm.DnsServer, v => v.DnsServerComboBox.Text)
                .DisposeWith(disposables);
            this.Bind(ViewModel, vm => vm.Domain, v => v.DomainTextBox.Text)
                .DisposeWith(disposables);
            this.Bind(ViewModel, vm => vm.RecordType, v => v.RecordTypeComboBox.Text)
                .DisposeWith(disposables);
            this.Bind(ViewModel, vm => vm.RecordClass, v => v.RecordClassComboBox.Text)
                .DisposeWith(disposables);
            this.Bind(ViewModel, vm => vm.Sni, v => v.SniTextBox.Text)
                .DisposeWith(disposables);
            this.Bind(ViewModel, vm => vm.ClientSubnet, v => v.ClientSubnetTextBox.Text)
                .DisposeWith(disposables);
            this.Bind(ViewModel, vm => vm.Proxy, v => v.ProxyTextBox.Text)
                .DisposeWith(disposables);

            // 命令与状态绑定
            this.BindCommand(ViewModel, vm => vm.QueryCommand, v => v.QueryButton)
                .DisposeWith(disposables);
            this.OneWayBind(ViewModel, vm => vm.CanQuery, v => v.QueryButton.IsEnabled)
                .DisposeWith(disposables);
            this.OneWayBind(ViewModel, vm => vm.IsBusy, v => v.IsBusyCheckBox.IsChecked)
                .DisposeWith(disposables);

            // 结果与错误显示
            this.OneWayBind(ViewModel, vm => vm.ResultLog, v => v.ResultTextBlock.Text)
                .DisposeWith(disposables);
            this.OneWayBind(ViewModel, vm => vm.Error, v => v.ErrorTextBlock.Text)
                .DisposeWith(disposables);

            // 结果变化时自动滚到底
            this.WhenAnyValue(v => v.ViewModel!.ResultLog)
                .WhereNotNull()
                .Throttle(TimeSpan.FromMilliseconds(10))
                .ObserveOn(RxApp.MainThreadScheduler)
                .Subscribe(_ => ResultScroll?.ScrollToEnd())
                .DisposeWith(disposables);
        });
    }
}