#pragma once

#include "DnsQueryTask.h"

#include <QObject>
#include <QtQml/qqmlregistration.h>
class QThread;

class DnsQuery : public QObject
{
    Q_OBJECT
    QML_ELEMENT

public:
    explicit DnsQuery(QObject *parent = nullptr);

public slots:
    void startQuery();

protected:
    DnsQueryTask *m_worker;

    QString m_server;
    QString m_domain;
    QString m_type;
    QString m_classType;
    QString m_socks5Server;
    bool m_busy = false;

signals:
    void startWorkerDnsRequestQuery(const QString &server, const QString &domain, const QString &type, const QString &classType);
    void startWorkerDnsRequestOverSocks5Query(const QString &socks5Server, const QString &server, const QString &domain, const QString &type, const QString &classType);
    void busyChanged();
    void queryFinished(const QString &hostname, const QJsonObject &result);
    void queryFailed(const QString &hostname, const QJsonObject &error);

private slots:
    void handleQueryFinished(const QString &hostname, const QJsonObject &result);
    void handleQueryFailed(const QString &hostname, const QJsonObject &error);

public:
    QString server() const;
    void setServer(const QString &newServer);

    QString domain() const;
    void setDomain(const QString &newDomain);

    QString type() const;
    void setType(const QString &newType);

    QString classType() const;
    void setClassType(const QString &newClassType);

    QString socks5Server() const;
    void setSocks5Server(const QString &newSocks5Server);

    bool busy() const;

signals:
    void serverChanged();
    void domainChanged();
    void typeChanged();
    void classTypeChanged();
    void socks5ServerChanged();

private:
    Q_PROPERTY(QString server READ server WRITE setServer NOTIFY serverChanged FINAL)
    Q_PROPERTY(QString domain READ domain WRITE setDomain NOTIFY domainChanged FINAL)
    Q_PROPERTY(QString type READ type WRITE setType NOTIFY typeChanged FINAL)
    Q_PROPERTY(QString classType READ classType WRITE setClassType NOTIFY classTypeChanged FINAL)
    Q_PROPERTY(QString socks5Server READ socks5Server WRITE setSocks5Server NOTIFY socks5ServerChanged FINAL)
    Q_PROPERTY(bool busy READ busy CONSTANT FINAL)
};
