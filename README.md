# Smartctl Monitor Container

## Overview

This Docker container is designed to periodically check the health of specified disk devices using `smartctl` from the `smartmontools` suite and send email alerts if any warnings or errors are detected. It's ideal for system administrators and users who need to monitor the health of disks on servers or workstations.

## Prerequisites

- Docker installed on your host machine.
- `smartctl` compatible disk devices.
- An email account for sending alerts.

## Configuration

The behavior of the container is configured through environment variables. Below is a list of the available environment variables and their purpose:

- `SMTP_SERVER`: SMTP server used to send email alerts (e.g., `smtp.gmail.com`).
- `SMTP_PORT`: Port number for the SMTP server (e.g., `587`).
- `SENDER_EMAIL`: Email address used to send alerts.
- `SENDER_PASSWORD`: Password for the sender email account. Consider using an app-specific password if available.
- `RECIPIENT_EMAIL`: Email address where alerts should be sent.
- `SMARTCTL_LOCATION`: Location of the `smartctl` binary inside the container (default is `/usr/sbin/smartctl`).
- `DEVICES`: Comma-separated list of devices to monitor (e.g., `/dev/sda,/dev/sdb`).
- `CHECK_INTERVAL`: Interval between checks, in Go's duration format (e.g., `1h` for one hour).

Additionally, an optional `ENVFILE` environment variable can be used to specify the path to a file containing environment variable definitions (in `KEY=value` format). This can be used to easily switch between different configurations or to keep sensitive information secure.

- `ENVFILE`: Path to a file containing environment variable definitions. This file should be in the format of `KEY=value` per line.

## Running the Container

To run the container with the necessary capabilities to access disk hardware, use the `--cap-add=SYS_RAWIO` flag and specify your configuration through environment variables:

```sh
docker run -d \
  --cap-add=SYS_RAWIO \
  -e SMTP_SERVER=smtp.gmail.com \
  -e SMTP_PORT=587 \
  -e SENDER_EMAIL=your_email@gmail.com \
  -e SENDER_PASSWORD=your_password \
  -e RECIPIENT_EMAIL=recipient_email@gmail.com \
  -e SMARTCTL_LOCATION=/usr/sbin/smartctl \
  -e DEVICES=/dev/sda,/dev/sdb \
  -e CHECK_INTERVAL=1h \
  smartctl-monitor
```

If you are using an `ENVFILE` to configure the environment variables, ensure it is accessible within your Docker container and specify its path using the `-e ENVFILE=/path/to/your/envfile` option.

## Security Considerations

- Handle the `SENDER_PASSWORD` environment variable cautiously. Avoid hard-coding sensitive information and consider using Docker secrets or other secure mechanisms for managing credentials.
- Using `--cap-add=SYS_RAWIO` grants the container the specific capabilities required for the operation of `smartctl`, minimizing the risk associated with running containers with broad privileges. It's a safer alternative to using the `--privileged` flag, as it restricts the container's capabilities to only those that are absolutely necessary.