# Deploy

A new deploy from scratch takes approximately 20 minutes.

All flags are optional. Configuration settings provided via flags will persist in later deployments unless explicitly overriden.

In order for flags to be parsed correctly, the name of your deployment should be placed at the end of your command.

## Custom Domains

|**Flag**|**Description**|**Environment Variable**|
|:-|:-|:-|
|`--domain value`|Domain to use as endpoint for Concourse web interface (eg: ci.myproject.com)|`DOMAIN`|

```sh
control-tower deploy --domain chimichanga.engineerbetter.com chimichanga
```

In the example above `control-tower` will search for a hosted zone that matches `chimichanga.engineerbetter.com` or `engineerbetter.com` and add a record to the longest match (`chimichanga.engineerbetter.com` in this example).

>The domain you provide must fall within a hosted zone in the Cloud DNS of the GCP project or route53 of the AWS account you are deploying to. For example, in our system tests we test this by delegating gcp.engineerbetter.com to our GCP project (our root domain is managed on another DNS server) then specifying something like control-tower.gcp.engineerbetter.com as the domain.

## Custom TLS Certificates

|**Flag**|**Description**|**Environment Variable**|
|:-|:-|:-|
|`--tls-cert value`|TLS cert to use with Concourse endpoint|`TLS_CERT`|
|`--tls-key value`|TLS private key to use with Concourse endpoint|`TLS_KEY`|

>By default `control-tower` will generate a self-signed cert using the given domain. If you'd like to provide your own certificate instead, pass the cert and private key as strings using the `--tls-cert` and `--tls-key` flags respectively. eg:

```sh
control-tower deploy \
  --domain chimichanga.engineerbetter.com \
  --tls-cert "$(cat chimichanga.engineerbetter.com.crt)" \
  --tls-key "$(cat chimichanga.engineerbetter.com.key)" \
  chimichanga
```

## Worker Configuration

|**Flag**|**Description**|**Environment Variable**|
|:-|:-|:-|
|`--workers value`|Number of Concourse worker instances to deploy (default: 1)|`WORKERS`|
|`--worker-type`|Specify a worker type for aws (m5, m5a, or m4) (default: "m4")|`WORKER_TYPE`|
|`--worker-size value`|Size of Concourse workers. See table below for sizes<br>(default: "xlarge")|`WORKER_SIZE`|

**`worker-type` is an AWS-specific option**

> AWS does not offer m5 or m5a instances in all regions, and even for regions that do offer m5 instances, not all zones within that region may offer them. To complicate matters further, each AWS account is assigned AWS zones at random - for instance, `eu-west-1a` for one account may be the same as `eu-west-1b` in another account. If m5s are available in your chosen region but _not_ the zone Control Tower has chosen, create a new deployment, this time specifying another `--zone`.

|--worker-size|AWS m4 Instance type|AWS m5 Instance type|AWS m5a Instance type|GCP Instance type|
|:-|:-|:-|:-|:-|
|medium|t3.medium|t3.medium||n1-standard-1|
|large |m4.large|m5.large|m5a.large|n1-standard-2|
|xlarge|m4.xlarge|m5.xlarge|m5a.xlarge|n1-standard-4|
|2xlarge|m4.2xlarge|m5.2xlarge|m5a.2xlarge|n1-standard-8|
|4xlarge|m4.4xlarge|m5.4xlarge|m5a.4xlarge|n1-standard-16|
|10xlarge|m4.10xlarge|||n1-standard-32|
|12xlarge||m5.12xlarge|m5a.12xlarge||
|16xlarge|m4.16xlarge|||n1-standard-64|
|24xlarge||m5.24xlarge|m5a.24xlarge||

## Web Configuration

|**Flag**|**Description**|**Environment Variable**|
|:-|:-|:-|
|`--web-size value`|Size of Concourse web node. See table below for sizes<br>(default: "small")|`WEB_SIZE`|

|--web-size|AWS Instance type|GCP Instance type|
|:-|:-|:-|
|small|t3.small|n1-standard-1|
|medium|t3.medium|n1-standard-2|
|large|t3.large|n1-standard-4|
|xlarge|t3.xlarge|n1-standard-8|
|2xlarge|t3.2xlarge|n1-standard-16|

## Database Configuration

|**Flag**|**Description**|**Environment Variable**|
|:-|:-|:-|
|`--db-size value`|Size of Concourse Postgres instance. See table below for sizes<br>(default: "small")|`DB_SIZE`|

>Note that when changing the database size on an existing control-tower deployment, the SQL instance will scaled by terraform resulting in approximately 3 minutes of downtime.

|--db-size|AWS Instance type|GCP Instance type|
|:-|:-|:-|
|small|db.t3.small|db-g1-small|
|medium|db.t3.medium|db-custom-2-4096|
|large|db.m4.large|db-custom-2-8192|
|xlarge|db.m4.xlarge|db-custom-4-16384|
|2xlarge|db.m4.2xlarge|db-custom-8-32768|
|4xlarge|db.m4.4xlarge|db-custom-16-65536|

## Global Resources

|**Flag**|**Description**|**Environment Variable**|
|:-|:-|:-|
|`--enable-global-resources`|Enable [Global Resources](https://concourse-ci.org/global-resources.html) in the Concourse cluster. Can be true/false. Default is false.|`ENABLE_GLOBAL_RESOURCES`|

## Whitelisting IPs

|**Flag**|**Description**|**Environment Variable**|
|:-|:-|:-|
|`--allow-ips value`|Comma separated list of IP addresses or CIDR ranges to allow access to. Not applied to future manual deploys unless this flag is provided again<br>(default: "0.0.0.0/0")|`ALLOW_IPS`|

> `allow-ips` governs what can access Concourse but not what can access the control plane (i.e. the BOSH director). The control plane will be restricted to the IP `control-tower deploy` was run from.

> This flag overwrites the allowed IPs on every deploy. This means deploying with `allow-ips` then deploying again without it will reset the allow list to `0.0.0.0/0`. The self-update pipeline will maintain the `allow-ips` of the most recent deploy.

## BitBucket Auth

|**Flag**|**Description**|**Environment Variable**|
|:-|:-|:-|
|`--bitbucket-auth-client-id value`|Client ID for a bitbucket OAuth application - Used for Bitbucket Auth|`BITBUCKET_AUTH_CLIENT_ID`|
|`--bitbucket-auth-client-secret value`|Client Secret for a bitbucket OAuth application - Used for Bitbucket Auth|`BITBUCKET_AUTH_CLIENT_SECRET`|

## GitHub Auth

|**Flag**|**Description**|**Environment Variable**|
|:-|:-|:-|
|`--github-auth-client-id value`|Client ID for a github OAuth application - Used for Github Auth|`GITHUB_AUTH_CLIENT_ID`|
|`--github-auth-client-secret value`|Client Secret for a github OAuth application - Used for Github Auth|`GITHUB_AUTH_CLIENT_SECRET`|

## Microsoft Auth

|**Flag**|**Description**|**Environment Variable**|
|:-|:-|:-|
|`--microsoft-auth-client-id value`|Client ID for a microsoft OAuth application - Used for Microsoft Auth|`MICROSOFT_AUTH_CLIENT_ID`|
|`--microsoft-auth-client-secret value`|Client Secret for a microsoft OAuth application - Used for Microsoft Auth|`MICROSOFT_AUTH_CLIENT_SECRET`|
|`--microsoft-auth-tenant value`|Tenant for a microsoft OAuth application - Used for Microsoft Auth|`MICROSOFT_AUTH_TENANT`|

## Custom Tagging

|**Flag**|**Description**|**Environment Variable**|
|:-|:-|:-|
|`--add-tag key=value`|Add a tag to the VMs that form your `control-tower` deployment. Can be used multiple times in a single `deploy` command||

## Volatile Lifecycle VMs

|**Flag**|**Description**|**Environment Variable**|
|:-|:-|:-|
|`--spot=value`|Use spot instances for workers. Can be true/false. Default is true|`SPOT`|
|`--preemptible=value`|Use preemptible instances for workers. Can be true/false. Default is true|`PREEMPTIBLE`|

> Control Tower uses spot/preemptible instances for workers by default as a cost saving measure. Users requiring lower risk may switch this feature off by setting --spot=false.

> Be aware the [preemptible instances](https://cloud.google.com/preemptible-vms/) _will_ go down at least once every 24 hours so deployments with only one worker _will_ experience downtime with this feature enabled. BOSH will ressurect falled workers automatically.

`spot` and `preemptible` are interchangeable so if either of them is set to false then interruptible instances will not be used regardless of your IaaS. i.e:

```sh
# Results in an AWS deployment using non-spot workers
control-tower deploy --spot=true --preemptible=false <your-project-name>
# Results in an AWS deployment using non-spot workers
control-tower deploy --preemptible=false <your-project-name>
# Results in a GCP deployment using non-preemptible workers
control-tower deploy --iaas gcp --spot=false <your-project-name>
```

## Availability Zone Selection

|**Flag**|**Description**|**Environment Variable**|
|:-|:-|:-|
|`--zone`|Specify an availability zone|`ZONE`|

> This cannot be changed after the initial deployment

## Custom CIDR ranges

If any of the following 5 flags is set, all the required ones from this group need to be set (The `rds` ones are AWS-Specific)

|**Flag**|**Description**|**Environment Variable**|
|:-|:-|:-|
|`--vpc-network-range value`|Customise the VPC network CIDR to deploy into<br>(required for AWS)|`VPC_NETWORK_RANGE`|
|`--public-subnet-range value`|Customise public network CIDR (if IAAS is AWS must be within --vpc-network-range)<br>(required)|`PUBLIC_SUBNET_RANGE`|
|`--private-subnet-range value`|Customise private network CIDR (if IAAS is AWS must be within --vpc-network-range)<br>(required)|`PRIVATE_SUBNET_RANGE`|
|`--rds-subnet-range1 value`|Customise first rds network CIDR (must be within --vpc-network-range)<br>(required for AWS)|`RDS_SUBNET_RANGE1`|
|`--rds-subnet-range2 value`|Customise second rds network CIDR (must be within --vpc-network-range)<br>(required for AWS)|`RDS_SUBNET_RANGE2`|

> All the ranges above should be in the CIDR format of IPv4/Mask. The sizes can vary as long as `vpc-network-range` is big enough to contain all others (in case IAAS is AWS). The smallest CIDR for `public` and `private` subnets is a /28. The smallest CIDR for `rds1` and `rds2` subnets is a /29

## Disable Colocated Metrics Stack

By default Control Tower colocates Grafana, Telegraf, and InfluxDB into the Concourse VMs. This can cause uneccessary resource usage if you don't use these features. It can be disabled with:

|**Flag**|**Description**|**Environment Variable**|
|:-|:-|:-|
|`--no-metrics`|Don't deploy the metrics stack colocated on the web VM (default: true)|`NO_METRICS`|

> In order to re-enable metrics after using this flag you need to deploy with `--no-metrics=false`.
