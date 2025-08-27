#include <QGuiApplication>
#include <QQmlApplicationEngine>
#include <QQuickStyle>
#include <QTranslator>

int main(int argc, char *argv[])
{
    QGuiApplication app(argc, argv);

    // 获取系统语言
    QString systemLanguage = QLocale::system().name();
    qDebug() << "System Language:" << systemLanguage;

    // 创建翻译器
    QTranslator translator;

    // 如果语言是中文
    if (systemLanguage == QStringLiteral("zh_CN"))
    {
        bool loaded = translator.load(":/NetTest_zh_CN.qm");
        qDebug() << "Translation loaded:" << loaded;

        if (loaded)
        {
            app.installTranslator(&translator);
        }
    }

    QQuickStyle::setStyle("Material");

    QQmlApplicationEngine engine;
    QObject::connect(
        &engine,
        &QQmlApplicationEngine::objectCreationFailed,
        &app,
        []()
    {
        QCoreApplication::exit(-1);
    },
        Qt::QueuedConnection);
    engine.loadFromModule("NetTest", "Main");

    return app.exec();
}
