using ReactiveUI.Maui;
using Service.ViewModel;

namespace MauiProgram.View;

public class DnsQueryPageBase : ReactiveContentPage<DnsQueryViewModel> { }

public partial class DnsQueryPage : DnsQueryPageBase
{
	public DnsQueryPage()
	{
		InitializeComponent();

        ViewModel = new DnsQueryViewModel();

    }
}