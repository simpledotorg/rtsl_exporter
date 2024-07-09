# RTSL Exporter for Prometheus

A Prometheus exporter that exposes multiple custom metrics.


> [!WARNING]  
> This project is for **testing and educational purposes only**. Comprehensive error handling and tests are intentionally omitted from this project. 
> Please **DO NOT** use this in production without adding thorough error checking, proper logging, and comprehensive unit tests. Using code without adequate error handling and tests can cause unexpected behavior and may lead to security vulnerabilities.

## Local setup

1. Clone this repository.

```sh
git clone https://github.com/simpledotorg/rtsl_exporter
```

2. Navigate to the project directory.

```sh
cd rtsl_exporter
```

3. Create and update config.yaml file
   
```sh
cp config.yaml.sample config.yaml
# Update the sample values
```

4. Build and Run

```sh
go build -o rtsl_exporter
./rtsl_exporter
```

Now, you can find your metrics at `http://localhost:8080/metrics`.

### Docker

You can build and run this exporter using Docker:

```sh
cp config.yaml.sample config.yaml
# Update the sample values

docker build -t rtsl_exporter .
docker run -p 8080:8080 -v ./config.yaml:/app/config.yaml rtsl_exporter
```

## Metrics

This exporter provides the following metrics:

- alphasms_user_balance_amount: The current balance amount
- alphasms_user_balance_error: The current error code while connecting to api
- alphasms_user_balance_validity: Validity date of balance amount
- dhis2_system_info_<domain name>: Information about the DHIS2 system

Exporter will also include exporter specific metrics
