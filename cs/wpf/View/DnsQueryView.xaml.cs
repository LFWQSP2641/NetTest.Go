using ReactiveUI;
using Service;
using Service.ViewModel;
using System.Collections.Specialized;
using System.Reactive.Disposables;
using System.Windows;
using System.Windows.Controls;
using System.Windows.Data;
using System.Windows.Media;

namespace wpf.View;

public class DnsQueryViewBase : ReactiveUI.ReactiveUserControl<DnsQueryViewModel> { }

public partial class DnsQueryView : DnsQueryViewBase
{
    public DnsQueryView()
    {
        InitializeComponent();
        ViewModel = new DnsQueryViewModel();

        // 立即配置结果列表，避免激活延迟导致初次不显示
        ResultItems.ItemsSource = ViewModel.Entries;
            ResultItems.ItemsPanel = new ItemsPanelTemplate(new FrameworkElementFactory(typeof(StackPanel)));
            ResultItems.ItemTemplate = BuildItemTemplate();

        DnsSchemeComboBox.ItemsSource = Global.DnsSchemes;
        DnsServerComboBox.ItemsSource = Global.CommonDnsServers;
        RecordTypeComboBox.ItemsSource = Global.DnsRecordType;
        RecordClassComboBox.ItemsSource = Global.DnsRecordClass;

        this.WhenActivated(disposables =>
        {
            this.Bind(ViewModel, vm => vm.DnsServerScheme, v => v.DnsSchemeComboBox.Text)
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

            // 按项展示：设置 ItemsSource 并监听集合变更时滚动到底部
            var entries = ViewModel!.Entries;
            // ItemsControl 配置与模板（代码构建）
            ResultItems.ItemsSource = entries;
            ResultItems.ItemsPanel = new ItemsPanelTemplate(new FrameworkElementFactory(typeof(StackPanel)));
            ResultItems.ItemTemplate = BuildItemTemplate();
            NotifyCollectionChangedEventHandler handler = (s, e) => ResultScroll?.ScrollToEnd();
            entries.CollectionChanged += handler;
            Disposable.Create(() => entries.CollectionChanged -= handler)
                .DisposeWith(disposables);
        });
    }

        private DataTemplate BuildItemTemplate()
        {
            var template = new DataTemplate { DataType = typeof(DnsQueryViewModel.LogEntry) };

            // Border -> StackPanel -> [header, body]
            var rootBorderFactory = new FrameworkElementFactory(typeof(Border));
            rootBorderFactory.SetValue(Border.PaddingProperty, new Thickness(8, 6, 8, 6));
            rootBorderFactory.SetValue(Border.BorderThicknessProperty, new Thickness(0));
            rootBorderFactory.SetValue(Border.SnapsToDevicePixelsProperty, true);

            var stackFactory = new FrameworkElementFactory(typeof(StackPanel));
            stackFactory.SetValue(StackPanel.OrientationProperty, Orientation.Vertical);

            // Header: [time] [type]
            var headerFactory = new FrameworkElementFactory(typeof(TextBlock));
            headerFactory.SetBinding(TextBlock.TextProperty, new Binding("Header"));
            headerFactory.SetValue(TextBlock.FontWeightProperty, FontWeights.SemiBold);
            headerFactory.SetValue(TextBlock.MarginProperty, new Thickness(0, 0, 0, 2));

            // Body: content, red if error
            var bodyFactory = new FrameworkElementFactory(typeof(TextBlock));
            bodyFactory.SetBinding(TextBlock.TextProperty, new Binding("Body"));
            bodyFactory.SetValue(TextBlock.TextWrappingProperty, TextWrapping.Wrap);
            bodyFactory.SetValue(TextBlock.MarginProperty, new Thickness(0, 0, 0, 6));

            var textStyle = new Style(typeof(TextBlock));
            var errorTrigger = new DataTrigger
            {
                Binding = new Binding("IsError"),
                Value = true
            };
            errorTrigger.Setters.Add(new Setter(TextBlock.ForegroundProperty, Brushes.IndianRed));
            textStyle.Triggers.Add(errorTrigger);
            bodyFactory.SetValue(TextBlock.StyleProperty, textStyle);

            stackFactory.AppendChild(headerFactory);
            stackFactory.AppendChild(bodyFactory);

            rootBorderFactory.AppendChild(stackFactory);

            template.VisualTree = rootBorderFactory;
            return template;
        }
}
