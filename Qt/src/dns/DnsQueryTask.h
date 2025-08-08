#pragma once

#include <QJsonObject>
#include <QObject>

class QLibrary;

class DnsQueryTask : public QObject
{
    Q_OBJECT

public:
    explicit DnsQueryTask(QObject *parent = nullptr);
    ~DnsQueryTask();

    bool isLoaded() const;

public slots:
    bool load();
    bool unload();
    QJsonObject dnsRequest(const QString &server, const QString &domain, const QString &type, const QString &classType);
    QJsonObject dnsRequestOverSocks5(const QString &socks5Server, const QString &server, const QString &domain, const QString &type, const QString &classType);

    void dnsRequestAsync(const QString &server, const QString &domain, const QString &type, const QString &classType);
    void dnsRequestOverSocks5Async(const QString &socks5Server, const QString &server, const QString &domain, const QString &type, const QString &classType);

protected:
    bool m_loaded = false;

    QLibrary *m_dnsLibrary;
    typedef const char *(*DnsRequest)(const char *, const char *, const char *, const char *);
    DnsRequest m_dnsRequest;
    typedef const char *(*DnsRequestOverSocks5)(const char *, const char *, const char *, const char *, const char *);
    DnsRequestOverSocks5 m_dnsRequestOverSocks5;
    typedef void (*FreeCString)(const char *);
    FreeCString m_freeCString;

    typedef void (*DnsCallback)(void *, char *);
    typedef const char *(*DnsRequestAsync)(const char *, const char *, const char *, const char *, DnsCallback, void *);
    DnsRequestAsync m_dnsRequestAsync;
    typedef const char *(*DnsRequestOverSocks5Async)(const char *, const char *, const char *, const char *, const char *, DnsCallback, void *);
    DnsRequestOverSocks5Async m_dnsRequestOverSocks5Async;

    static void dnsCallback(void *context, char *response);

protected slots:
    void freeCString(const char *str);
    QJsonObject handleDnsResponse(const QByteArray &response);
    std::optional<QString> funcPointerCheck();

signals:
    void loadFinished();
    void unloadFinished();
    void queryFinished(const QString &hostname, const QJsonObject &result);
    void queryFailed(const QString &hostname, const QJsonObject &error);

private:
    Q_PROPERTY(bool isLoaded READ isLoaded CONSTANT FINAL)
};
