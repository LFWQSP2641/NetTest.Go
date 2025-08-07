#include <QCoreApplication>
#include <QDebug>
#include <QLibrary>

int main(int argc, char *argv[])
{
    QCoreApplication a(argc, argv);

    auto lib = new QLibrary("netcore", &a);
    if (!lib->load())
    {
        qDebug() << "Failed to load library:" << lib->errorString();
        return -1;
    }
    typedef const char *(*DnsRequest)(const char *, const char *, const char *, const char *);
    DnsRequest dnsRequest = (DnsRequest)lib->resolve("DnsRequest");
    if (!dnsRequest)
    {
        qDebug() << "Failed to resolve function:" << lib->errorString();
        return -1;
    }

    typedef const char *(*DnsRequestOverSocks5)(const char *, const char *, const char *, const char *, const char *);
    DnsRequestOverSocks5 dnsRequestOverSocks5 = (DnsRequestOverSocks5)lib->resolve("DnsRequestOverSocks5");
    if (!dnsRequestOverSocks5)
    {
        qDebug() << "Failed to resolve function:" << lib->errorString();
        return -1;
    }

    typedef void (*FreeCString)(const char *);
    FreeCString freeCString = (FreeCString)lib->resolve("FreeCString");
    if (!freeCString)
    {
        qDebug() << "Failed to resolve function:" << lib->errorString();
        return -1;
    }

    {
        const char *result = dnsRequest("223.5.5.5:53", "dns.alidns.com", "A", "IN");
        if (result)
        {
            qDebug() << "Result length:" << strlen(result);
            qDebug() << "Result:" << result;
            freeCString(result);
        }
        else
        {
            qDebug() << "No result returned.";
        }
    }

    {
        const char *result = dnsRequestOverSocks5("192.168.100.1:20811", "223.5.5.5:53", "dns.alidns.com", "A", "IN");
        if (result)
        {
            qDebug() << "Result length:" << strlen(result);
            qDebug() << "Result:" << result;
            freeCString(result);
        }
        else
        {
            qDebug() << "No result returned.";
        }
    }

    return 0;
}
