# Portwhine

<div align="center">
    <img src="/assets/images/logo.png" alt="Logo" width="250">
</div>

Portwhine is a software for automatically checking assets, especially on the web. The idea is that there are certain input triggers which then trigger a check by various tools. These tools all run in docker containers and are started up and shut down on the fly as required. Results are stored in an elastic search database and can be analyzed via [kibana](https://www.elastic.co/kibana). There is also an API that can be used to make configurations.

## 📋 Table of Contents

- [🚀 Quick Start](#quick-start)
- [📖 Usage](#usage)
- [✨ Features](#features)
- [🛠️ Development](#development)

## 🚀 Quick Start
<a name="quick-start"></a>

The entire program can be built and started using the `make start` command. This command builds all necessary containers and then starts a docker compose file which starts the necessary databases, frontend and api services. The Makefile only works on Linux, but Windows users can execute the corresponding commands manually.

Before all services start up, some environment variables should be set. This can be done, for example, with an `.env` file in the root of this repository. The file could look like this:

```dotenv
# Optional: Service-Namen (Defaults sind in docker-compose.yml hinterlegt)
ES_HOST=elasticsearch
KIBANA_HOST=kibana
METRICBEAT_HOST=metricbeat
API_HOST=api
MINIO_HOST=minio
REDIS_HOST=redis
FRONTEND_HOST=frontend
CLIENT_CERT_NAME=client
TRAEFIK_HOST=traefik

# Wichtig für Worker-Container (API startet Worker via Docker-Socket):
# Muss ein HOST-Pfad zum lokalen certs-Ordner sein (z.B. Windows: C:\Users\...\Portwhine\certs)
HOST_CERTS_PATH=/absolute/path/to/Portwhine/certs

ELASTIC_USERNAME=elastic
ELASTIC_PASSWORD=changeme

APP_DB_USER=app_user
APP_DB_PASSWORD=app_password

APP_MINIO_USER=app_minio
APP_MINIO_PASSWORD=app_minio_password

MINIO_ROOT_USER=minio
MINIO_ROOT_PASSWORD=changeme

# Redis (default user is disabled for security)
REDIS_USER=app_redis
REDIS_PASSWORD=app_redis_password
```

Once all containers have been built and started, various services can be accessed via a browser. The frontend can be reached under `localhost:8443`. Kibana can be reached under `localhost:5601`. However, the ports and other configurations can be adjusted as required in the [docker compose](./docker-compose.yml).

## 📖 Usage
<a name="usage"></a>



## ✨ Features
<a name="features"></a>

The features of Portwhine are limitless. Joking aside, because Portwhine is only a platform for individual containers, any checks can be poured into containers and executed in a pipeline. The following checks are currently configurable:

### Trigger

| Name | Description | Settings |
|---|---|---|
| IPAddressTrigger | Trigger that accepts a list of IP addresses, single IP address, a list of networks, or single network. The repetition (seconds) defines if the trigger should start a scan repetitively. | `ip_addresses`; `repetition` |
| CertStreamTrigger | Trigger that accepts a regex pattern. The trigger monitors certificate transparency logs for new certificates that match the regex pattern. | `regex` |

### Worker

| Name | Description | Settings |
|---|---|---|
| NmapWorker | Worker that performs network mapping on IP addresses. |   |
| FFUFWorker | Worker that performs fuzzing on HTTP endpoints. |   |
| HumbleWorker | Worker that analyzes HTTP headers. |   |
| ResolverWorker | Worker that resolves domain names to IP addresses. | `use_internal` |
| ScreenshotWorker | Worker that takes screenshots of HTTP endpoints. |   |
| TestSSLWorker | Worker that tests SSL configurations on IP addresses. |   |
| WebAppAnalyzerWorker | Worker that analyzes web applications on HTTP endpoints. |   |

## Development
<a name="development"></a>

> :warning: **Development ongoing**: This software is currently under development and both the configurations and the APIs are subject to change.

To develop new modules for portwhine, not many conditions need to be met. Modules are free to access the database or start other modules. however, to interact with the api, a certain format must be followed. This is defined in [job_payload](https://github.com/morzan1001/Portwhine/blob/main/src/modules/api/models/job_payload.py). Containers can write results to the database and then notify the api via an http call.

To make changes to the API or other containers, for example, these must be rebuilt. To avoid having to constantly rebuild all containers during debugging, here are a few practical commands:

```bash
docker build --no-cache -f docker/Dockerfile.api -t api:1.0 .
```

```bash
docker compose create api
```

```bash
docker compose start api
```

I am happy if you contribute changes or new modules in the form of a pull request 😃

### Future Plans

In the future i would like to focus on providing new containers and modules for portwhine. Below is a list of modules i will implement next.

- [wafw00f](https://github.com/EnableSecurity/wafw00f)
- [MANSPIDER](https://github.com/blacklanternsecurity/MANSPIDER)

It would also be good to have a feature that can generate reports and then send them by e-mail.