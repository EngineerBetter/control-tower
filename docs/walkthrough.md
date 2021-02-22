# What Control Tower does

`control-tower` first creates an S3 or GCS bucket to store its own configuration and saves a `config.json` file there.

It then uses Terraform to deploy the following infrastructure:

- AWS
  - Key pair
  - S3 bucket for the blobstore
  - IAM user that can access the blobstore
    - IAM access key
    - IAM user policy
  - IAM user that can deploy EC2 instances
    - IAM access key
    - IAM user policy
  - VPC
  - Internet gateway
  - Route for internet_access
  - NAT gateway
  - Route table for private
  - Subnet for public
  - Subnet for private
  - Route table association for private
  - Route53 record for Concourse
  - EIP for director, ATC, and NAT
  - Security groups for director, vms, RDS, and ATC
  - Route table for RDS
  - Route table associations for RDS
  - Subnets for RDS
  - DB subnet group
  - DB instance
- GCP
  - A DNS A record pointing to the ATC IP
  - A Compute route for the nat instance
  - A Compute instance for the nat
  - A Compute network
  - Public and Private Compute subnetworks
  - Compute firewalls for director, nat, atc-one, atc-two, vms, atc-three, internal, and sql
  - A Service account for for bosh
  - A Service account key for bosh
  - A Project iam member for bosh
  - Compute addresses for the ATC and Director
  - A Sql database instance
  - A Sql database
  - A Sql user

Once the terraform step is complete, `control-tower` deploys a BOSH director on an t3.small/n1-standard-1 instance, and then uses that to deploy a Concourse with the following settings:

- One t3.small/n1-standard-1 for the Concourse web server
- One m4.xlarge [spot](https://aws.amazon.com/ec2/spot/)/n1-standard-4 [preemptible](https://cloud.google.com/preemptible-vms/) instance used as a Concourse worker
- Access via over HTTP and HTTPS using a user-provided certificate, or an auto-generated self-signed certificate if one isn't provided.
