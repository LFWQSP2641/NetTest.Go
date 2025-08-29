using MauiProgram.View;
using ReactiveUI;
using Service.ViewModel;
using Splat;

namespace MauiProgram
{
    public static class MauiProgram
    {
        public static MauiApp CreateMauiApp()
        {
            var builder = MauiApp.CreateBuilder();
            builder
                .UseMauiApp<App>();

            builder.Services.AddTransient<DnsQueryPage>();
            builder.Services.AddScoped<DnsQueryViewModel>();

            return builder.Build();
        }
    }
}
