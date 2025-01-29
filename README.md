# Portwhine

<div align="center">
    <img src="/assets/images/logo.png" alt="Logo" width="250">
</div>

Portwhine is a software for automatically checking assets, especially on the web. The idea is that there are certain input triggers which then trigger a check by various tools. These tools all run in docker containers and are started up and shut down on the fly as required. Results are stored in an elastic search database and can be analyzed via kibana. There is also an API that can be used to make configurations.

## 📋 Table of Contents

- [🚀 Quick Start](#quick-start)
- [✨ Features](#features)
- [🛠️ Development](#development)

## 🚀 Quick Start
<a name="quick-start"></a>

The entire program can be built and started using the `make start` command.

## ✨ Features
<a name="features"></a>

The features of Portwhine are limitless. joking aside, because Portwhine is only a platform for individual containers, any checks can be poured into containers and executed in a pipeline. The following checks are currently configurable:

### Trigger

| Name | Description | Settings |
|---|---|---|
| IPAddressTrigger |   |   |
| CertStreamTrigger |   |   |

### Worker

| Name | Description | Settings |
|---|---|---|
| NmapWorker |   |   |
| FFUFWorker |   |   |
| HumbleWorker |   |   |
| ResolverWorker |   |   |
| ScreenshotWorker |   |   |
| TestSSLWorker |   |   |
| WebAppAnalyzerWorker |   |   |

## Development
<a name="development"></a>

> :warning: **Development ongoing**: This software is currently under development and both the configurations and the APIs are subject to change.

### Future Plans

https://github.com/EnableSecurity/wafw00f
https://github.com/blacklanternsecurity/MANSPIDER

