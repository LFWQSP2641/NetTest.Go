using Avalonia.ReactiveUI;
using Avalonia.Threading;
using ReactiveUI;
using Service;
using Service.ViewModel;
using System.Reactive.Disposables;
using System.Reactive.Linq;
using System.Collections.Specialized;
using Avalonia.Controls;
using Avalonia.Controls.Templates;
using Avalonia.Media;
using Avalonia;

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

            // 错误显示
            this.OneWayBind(ViewModel, vm => vm.Error, v => v.ErrorTextBlock.Text)
                .DisposeWith(disposables);

            // 按项展示：设置 ItemsSource 与模板（代码中构建），并监听集合变更滚动到底部
            var entries = ViewModel!.Entries;
            ResultList.ItemsSource = entries;

            ResultList.ItemTemplate = new FuncDataTemplate<Service.ViewModel.DnsQueryViewModel.LogEntry>((item, ns) =>
            {
                var border = new Border
                {
                    CornerRadius = new CornerRadius(4),
                    Padding = new Thickness(8),
                    Margin = new Thickness(0, 0, 0, 8)
                };

                var header = new TextBlock { FontWeight = Avalonia.Media.FontWeight.Bold };
                header.Bind(TextBlock.TextProperty, new Avalonia.Data.Binding("Header"));

                var body = new TextBlock { TextWrapping = TextWrapping.Wrap };
                body.Bind(TextBlock.TextProperty, new Avalonia.Data.Binding("Body"));

                // 错误项红色（在 DataContext 变化后根据 IsError 设置颜色）
                border.AttachedToVisualTree += (_, __) =>
                {
                    if (border.DataContext is Service.ViewModel.DnsQueryViewModel.LogEntry le)
                    {
                        if (le.IsError)
                        {
                            header.Foreground = Brushes.IndianRed;
                            body.Foreground = Brushes.IndianRed;
                        }
                        else
                        {
                            // 使用默认前景（跟随主题），因此清除前景值
                            header.ClearValue(TextBlock.ForegroundProperty);
                            body.ClearValue(TextBlock.ForegroundProperty);
                        }
                    }
                };

                var stack = new StackPanel { Spacing = 4 };
                stack.Children.Add(header);
                stack.Children.Add(body);
                border.Child = stack;

                return border;
            }, true);

            NotifyCollectionChangedEventHandler handler = (s, e) => ResultScroll?.ScrollToEnd();
            entries.CollectionChanged += handler;
            Disposable.Create(() => entries.CollectionChanged -= handler)
                .DisposeWith(disposables);
        });
    }
}