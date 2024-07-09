# AlphaSMS Exporter for Prometheus

A Prometheus exporter that exposes metrics from the AlphaSMS API.


> [!WARNING]  
> This project is for **testing and educational purposes only**. Comprehensive error handling and tests are intentionally omitted from this project. 
> Please **DO NOT** use this in production without adding thorough error checking, proper logging, and comprehensive unit tests. Using code without adequate error handling and tests can cause unexpected behavior and may lead to security vulnerabilities.

## Local setup

1. Clone this repository.

```sh
git clone https://github.com/yourusername/alphasms_exporter
```

2. Navigate to the project directory.

```sh
cd alphasms_exporter
```

3. Set the AlphaSMS API key as an environment variable.
   Ensure to replace `your-api-key` with your actual API key.
   
```sh
export ALPHASMS_API_KEY=your-api-key
```

4. Build and Run

```sh
go build -o alphasms_exporter
./alphasms_exporter
```

Now, you can find your metrics at `http://localhost:8080/metrics`.

### Docker

You can build and run this exporter using Docker:

```sh
docker build -t alphasms_exporter .
docker run -p 8080:8080  -e ALPHASMS_API_KEY='your-api-key' alphasms_exporter
```

Replace `your-api-key` with your actual API key.

## Metrics

This exporter provides the following metrics:

- alphasms_user_balance_amount: The current balance amount
- alphasms_user_balance_error: The current error code while connecting to api
- alphasms_user_balance_validity: Validity date of balance amount

Exporter will also include exporter specific metrics

## Adding More Metrics

To expose more metrics from the AlphaSMS API, you can:

1. Define new methods similar to `GetUserBalance` in `alphasms/client.go`

2. In `alphasms/exporter.go`, define new Prometheus metrics and update the `Describe` and `Collect` methods accordingly.

3. Add the invocation of your new methods to the `Collect` method.
