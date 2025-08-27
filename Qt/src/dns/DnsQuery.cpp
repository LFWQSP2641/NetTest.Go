#include "DnsQuery.h"

#include "DnsQueryTask.h"

#include <QEventLoop>
#include <QThread>

DnsQuery::DnsQuery(QObject *parent)
    : QObject {parent},
      m_worker(new DnsQueryTask(this))
{
    connect(m_worker, &DnsQueryTask::queryFinished, this, &DnsQuery::handleQueryFinished);
    connect(m_worker, &DnsQueryTask::queryFailed, this, &DnsQuery::handleQueryFailed);
    connect(this, &DnsQuery::startWorkerDnsRequestQuery, m_worker, &DnsQueryTask::dnsRequestAsync);
    connect(this, &DnsQuery::startWorkerDnsRequestOverSocks5Query, m_worker, &DnsQueryTask::dnsRequestOverSocks5Async);
}

void DnsQuery::startQuery()
{
    if (m_busy)
        return;

    m_busy = true;
    emit busyChanged();

    if (!m_worker->isLoaded())
    {
        m_worker->load();
    }

    auto t_server = m_server.trimmed();
    auto t_domain = m_domain.trimmed();
    auto t_type = m_type.trimmed();
    auto t_classType = m_classType.trimmed();
    auto t_socks5Server = m_socks5Server.trimmed();
    auto t_sni = m_sni.trimmed();
    auto t_clientSubnet = m_clientSubnet.trimmed();

    // Normalize port and path by scheme
    auto ensurePort = [](QString &url, const QString &defaultPort)
    {
        const int colonIndex = url.lastIndexOf(':');
        if (colonIndex == -1)
        {
            url += defaultPort;
            return;
        }
        bool isNumber = false;
        const int port = QStringView {url}.mid(colonIndex + 1).toInt(&isNumber);
        if (!isNumber || port < 1 || port > 65535) url += defaultPort;
    };

    if (t_server.startsWith(QStringLiteral("tls://")) ||
        t_server.startsWith(QStringLiteral("quic://")) ||
        t_server.startsWith(QStringLiteral("doq://")))
    {
        ensurePort(t_server, QStringLiteral(":853"));
    }
    else if (t_server.startsWith(QStringLiteral("https://")) ||
             t_server.startsWith(QStringLiteral("https3://")) ||
             t_server.startsWith(QStringLiteral("http3://")) ||
             t_server.startsWith(QStringLiteral("h3://")))
    {
        // HTTPS/HTTP3: no port normalization, but ensure path
        const int schemeEnd = t_server.indexOf(QLatin1String("://"));
        const int hostStart = (schemeEnd >= 0) ? schemeEnd + 3 : 0;
        const int slashPos = t_server.indexOf('/', hostStart);
        if (slashPos < 0) t_server += QStringLiteral("/dns-query");
    }
    else
    {
        ensurePort(t_server, QStringLiteral(":53"));
    }
    if (t_type.isEmpty())
    {
        t_type = "A";
    }
    if (t_classType.isEmpty())
    {
        t_classType = "IN";
    }

    if (!t_socks5Server.isEmpty())
    {
        emit startWorkerDnsRequestOverSocks5Query(t_socks5Server, t_server, t_domain, t_type, t_classType, t_sni, t_clientSubnet);
    }
    else
    {
        emit startWorkerDnsRequestQuery(t_server, t_domain, t_type, t_classType, t_sni, t_clientSubnet);
    }
}

void DnsQuery::handleQueryFinished(const QString &hostname, const QJsonObject &result)
{
    m_busy = false;
    emit busyChanged();
    emit queryFinished(hostname, QJsonDocument(result).toJson());
}

void DnsQuery::handleQueryFailed(const QString &hostname, const QJsonObject &error)
{
    m_busy = false;
    emit busyChanged();
    emit queryFailed(hostname, QJsonDocument(error).toJson());
}

QString DnsQuery::server() const
{
    return m_server;
}

void DnsQuery::setServer(const QString &newServer)
{
    if (m_server == newServer)
        return;
    m_server = newServer;
    emit serverChanged();
}

QString DnsQuery::domain() const
{
    return m_domain;
}

void DnsQuery::setDomain(const QString &newDomain)
{
    if (m_domain == newDomain)
        return;
    m_domain = newDomain;
    emit domainChanged();
}

QString DnsQuery::type() const
{
    return m_type;
}

void DnsQuery::setType(const QString &newType)
{
    if (m_type == newType)
        return;
    m_type = newType;
    emit typeChanged();
}

QString DnsQuery::classType() const
{
    return m_classType;
}

void DnsQuery::setClassType(const QString &newClassType)
{
    if (m_classType == newClassType)
        return;
    m_classType = newClassType;
    emit classTypeChanged();
}

QString DnsQuery::socks5Server() const
{
    return m_socks5Server;
}

void DnsQuery::setSocks5Server(const QString &newSocks5Server)
{
    if (m_socks5Server == newSocks5Server)
        return;
    m_socks5Server = newSocks5Server;
    emit socks5ServerChanged();
}

QString DnsQuery::clientSubnet() const
{
    return m_clientSubnet;
}

void DnsQuery::setClientSubnet(const QString &newClientSubnet)
{
    if (m_clientSubnet == newClientSubnet)
        return;
    m_clientSubnet = newClientSubnet;
    emit clientSubnetChanged();
}

QString DnsQuery::sni() const
{
    return m_sni;
}

void DnsQuery::setSni(const QString &newSni)
{
    if (m_sni == newSni)
        return;
    m_sni = newSni;
    emit sniChanged();
}

bool DnsQuery::busy() const
{
    return m_busy;
}
