#include "DnsQueryTask.h"

#include <QDebug>
#include <QLibrary>

DnsQueryTask::DnsQueryTask(QObject *parent)
    : QObject {parent},
      m_dnsLibrary(nullptr),
      m_dnsRequest(nullptr),
      m_dnsRequestOverSocks5(nullptr),
      m_freeCString(nullptr),
      m_dnsRequestAsync(nullptr),
      m_dnsRequestOverSocks5Async(nullptr)
{
}

DnsQueryTask::~DnsQueryTask()
{
    if (m_dnsLibrary && m_dnsLibrary->isLoaded())
    {
        m_dnsLibrary->unload();
    }
}

bool DnsQueryTask::load()
{
#ifdef Q_OS_ANDROID
    const auto libName = "libandroidnetcore";
#else
    const auto libName = "netcore";
#endif
    m_dnsLibrary = new QLibrary(libName, this);
    if (!m_dnsLibrary->load())
    {
        qWarning() << "Failed to load library:" << m_dnsLibrary->errorString();
        return false;
    }

    m_dnsRequest = reinterpret_cast<DnsRequest>(m_dnsLibrary->resolve("DnsRequest"));
    if (!m_dnsRequest)
    {
        qWarning() << "Failed to resolve DnsRequest function:" << m_dnsLibrary->errorString();
        return false;
    }

    m_dnsRequestOverSocks5 = reinterpret_cast<DnsRequestOverSocks5>(m_dnsLibrary->resolve("DnsRequestOverSocks5"));
    if (!m_dnsRequestOverSocks5)
    {
        qWarning() << "Failed to resolve DnsRequestOverSocks5 function:" << m_dnsLibrary->errorString();
        return false;
    }

    m_freeCString = reinterpret_cast<FreeCString>(m_dnsLibrary->resolve("FreeCString"));
    if (!m_freeCString)
    {
        qWarning() << "Failed to resolve FreeCString function:" << m_dnsLibrary->errorString();
        return false;
    }

    m_dnsRequestAsync = reinterpret_cast<DnsRequestAsync>(m_dnsLibrary->resolve("DnsRequestAsync"));
    if (!m_dnsRequestAsync)
    {
        qWarning() << "Failed to resolve DnsRequestAsync function:" << m_dnsLibrary->errorString();
        return false;
    }

    m_dnsRequestOverSocks5Async = reinterpret_cast<DnsRequestOverSocks5Async>(m_dnsLibrary->resolve("DnsRequestOverSocks5Async"));
    if (!m_dnsRequestOverSocks5Async)
    {
        qWarning() << "Failed to resolve DnsRequestOverSocks5Async function:" << m_dnsLibrary->errorString();
        return false;
    }

    m_loaded = true;
    emit loadFinished();
    return true;
}

bool DnsQueryTask::unload()
{
    if (m_dnsLibrary && m_dnsLibrary->isLoaded())
    {
        m_dnsLibrary->unload();
        m_dnsLibrary = nullptr;
        m_dnsRequest = nullptr;
        m_dnsRequestOverSocks5 = nullptr;
        m_freeCString = nullptr;
        m_dnsRequestAsync = nullptr;
        m_dnsRequestOverSocks5Async = nullptr;
        m_loaded = false;
        emit unloadFinished();
        return true;
    }
    return false;
}

QJsonObject DnsQueryTask::dnsRequest(const QString &server, const QString &domain, const QString &type, const QString &classType)
{
    const auto err = funcPointerCheck();
    if (err)
    {
        return QJsonObject {
            {"code",    -1         },
            {"message", err.value()}
        };
    }
    auto result = m_dnsRequest(server.toUtf8().constData(),
                               domain.toUtf8().constData(),
                               type.toUtf8().constData(),
                               classType.toUtf8().constData());
    if (result)
    {
        const auto jsonResponse = handleDnsResponse(QByteArray(result));
        freeCString(result);
        emit queryFinished(domain, jsonResponse);
        return jsonResponse;
    }
    else
    {
        const QJsonObject error {
            {"code",    -2                   },
            {"message", "No result returned."}
        };
        emit queryFailed(domain, error);
        return error;
    }
}

QJsonObject DnsQueryTask::dnsRequestOverSocks5(const QString &socks5Server, const QString &server, const QString &domain, const QString &type, const QString &classType)
{
    const auto err = funcPointerCheck();
    if (err)
    {
        return QJsonObject {
            {"code",    -1         },
            {"message", err.value()}
        };
    }
    auto result = m_dnsRequestOverSocks5(socks5Server.toUtf8().constData(),
                                         server.toUtf8().constData(),
                                         domain.toUtf8().constData(),
                                         type.toUtf8().constData(),
                                         classType.toUtf8().constData());
    if (result)
    {
        const auto jsonResponse = handleDnsResponse(QByteArray(result));
        freeCString(result);
        emit queryFinished(domain, jsonResponse);
        return jsonResponse;
    }
    else
    {
        const QJsonObject error {
            {"code",    -2                   },
            {"message", "No result returned."}
        };
        emit queryFailed(domain, error);
        return error;
    }
}

void DnsQueryTask::dnsRequestAsync(const QString &server, const QString &domain, const QString &type, const QString &classType)
{
    const auto err = funcPointerCheck();
    if (err)
    {
        emit queryFailed(domain, QJsonObject {
                                     {"code",    -1         },
                                     {"message", err.value()}
        });
        return;
    }

    m_dnsRequestAsync(server.toUtf8().constData(),
                      domain.toUtf8().constData(),
                      type.toUtf8().constData(),
                      classType.toUtf8().constData(),
                      &DnsQueryTask::dnsCallback,
                      this);
}

void DnsQueryTask::dnsRequestOverSocks5Async(const QString &socks5Server, const QString &server, const QString &domain, const QString &type, const QString &classType)
{
    const auto err = funcPointerCheck();
    if (err)
    {
        emit queryFailed(domain, QJsonObject {
                                     {"code",    -1         },
                                     {"message", err.value()}
        });
        return;
    }

    m_dnsRequestOverSocks5Async(socks5Server.toUtf8().constData(),
                                server.toUtf8().constData(),
                                domain.toUtf8().constData(),
                                type.toUtf8().constData(),
                                classType.toUtf8().constData(),
                                &DnsQueryTask::dnsCallback,
                                this);
}

bool DnsQueryTask::isLoaded() const
{
    return m_loaded;
}

void DnsQueryTask::dnsCallback(void *context, char *response)
{
    if (!context || !response)
    {
        qWarning() << "DnsCallback received null context or response.";
        return;
    }

    DnsQueryTask *task = static_cast<DnsQueryTask *>(context);

    QJsonObject jsonResponse = task->handleDnsResponse(QByteArray(response));
    if (jsonResponse.contains("code") && jsonResponse["code"].toInt() < 0)
    {
        emit task->queryFailed(QString::fromUtf8(response), jsonResponse);
    }
    else
    {
        emit task->queryFinished(QString::fromUtf8(response), jsonResponse);
    }
}

void DnsQueryTask::freeCString(const char *str)
{
    const auto err = funcPointerCheck();
    if (err)
    {
        qWarning() << "Function pointer check failed:" << err.value();
        return;
    }
    if (str && m_freeCString)
    {
        m_freeCString(str);
    }
}

QJsonObject DnsQueryTask::handleDnsResponse(const QByteArray &response)
{
    qDebug() << response;
    QJsonObject jsonResponse = QJsonDocument::fromJson(response).object();
    qDebug() << jsonResponse;
    if (jsonResponse.isEmpty())
    {
        jsonResponse = QJsonObject {
            {"code",    -3                     },
            {"message", "Invalid DNS response."}
        };
    }
    return jsonResponse;
}

std::optional<QString> DnsQueryTask::funcPointerCheck()
{
    if (!m_dnsLibrary || !m_dnsLibrary->isLoaded())
    {
        return QStringLiteral("DNS library is not loaded.");
    }
    if (!m_dnsRequest)
    {
        return QStringLiteral("DnsRequest function pointer is null.");
    }
    if (!m_dnsRequestOverSocks5)
    {
        return QStringLiteral("DnsRequestOverSocks5 function pointer is null.");
    }
    if (!m_freeCString)
    {
        return QStringLiteral("FreeCString function pointer is null.");
    }
    if (!m_dnsRequestAsync)
    {
        return QStringLiteral("DnsRequestAsync function pointer is null.");
    }
    if (!m_dnsRequestOverSocks5Async)
    {
        return QStringLiteral("DnsRequestOverSocks5Async function pointer is null.");
    }
    return std::nullopt;
}
