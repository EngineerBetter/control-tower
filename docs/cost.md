# Estimated Cost

By default, `control-tower` deploys to the AWS eu-west-1 (Ireland) region or the GCP europe-west1 (Belgium) region, and uses spot instances for large and xlarge Concourse VMs. The estimated monthly cost is as follows:

## AWS

| Component     | Size             | Count | Price (USD) |
|---------------|------------------|-------|------------:|
| BOSH director | t2.small         |     1 |       18.30 |
| Web Server    | t2.small         |     1 |       18.30 |
| Worker        | m4.xlarge (spot) |     1 |      ~50.00 |
| RDS instance  | db.t2.small      |     1 |       28.47 |
| NAT Gateway   |         -        |     1 |       35.15 |
| gp2 storage   | 20GB (bosh, web) |     2 |        4.40 |
| gp2 storage   | 200GB (worker)   |     1 |       22.00 |
| **Total**     |                  |       |  **176.62** |

## GCP

| Component     | Size                              | Count | Price (USD) |
|---------------|-----------------------------------|-------|------------:|
| BOSH director | n1-standard-1                     |     1 |       26.73 |
| Web Server    | n1-standard-1                     |     1 |       26.73 |
| Worker        | n1-standard-4 (preemptible)       |     1 |       32.12 |
| DB instance   | db-g1-small                       |     1 |       27.25 |
| NAT Gateway   | n1-standard-1                     |     1 |       26.73 |
| disk storage  | 20GB (bosh, web) + 200GB (worker) |   -   |       40.80 |
| **Total**     |                                   |       |  **180.35** |
