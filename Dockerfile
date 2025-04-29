## Build stage
FROM public.ecr.aws/amazonlinux/amazonlinux:2 AS builder

RUN yum update -y && yum install -y golang git make

WORKDIR /app
COPY go.mod ./
RUN go mod download

COPY . .
RUN go build -o /app/BlockchainClient .

## Deployment stage
FROM public.ecr.aws/amazonlinux/amazonlinux:2

RUN yum update -y && yum install -y ca-certificates shadow-utils curl && \
    yum clean all && rm -rf /var/cache/yum

RUN useradd -m clientuser
USER clientuser
WORKDIR /home/clientuser
COPY --from=builder /app/BlockchainClient ./BlockchainClient
EXPOSE 8080
ENTRYPOINT ["./BlockchainClient"]