#!/usr/bin/env ruby

require 'json'

credentials = JSON.parse(File.read(ENV['GOOGLE_APPLICATION_CREDENTIALS']))

gcp_project_id = credentials['project_id']
account = credentials['client_email']
`gcloud config set project #{gcp_project_id}`
`gcloud auth activate-service-account #{account} --key-file=google_credentials.json`

unused_disks = JSON.parse(`gcloud compute disks list --filter="NOT users:*" --format=json`)

unused_disks.each do |disk|
  disk_name = disk['name']
  zone = disk['zone'].split('/').last
  puts "Deleting unused disk #{disk_name} in zone #{zone}..."
  `gcloud compute disks delete #{disk_name} --zone #{zone} --quiet`
end
