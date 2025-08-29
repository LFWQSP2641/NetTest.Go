using ReactiveUI;
using ReactiveUI.Maui;

namespace MauiProgram;

public class MainPageBase : ReactiveContentPage<ReactiveObject> { }

public partial class MainPage : MainPageBase
{
    public MainPage()
    {
        InitializeComponent();

        ViewModel = new ReactiveObject();
    }
}
