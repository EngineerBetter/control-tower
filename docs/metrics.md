# Metrics

Control Tower now automatically deploys Influxdb, Riemann, and Grafana on the web node. You can access Grafana on port 3000 of your regular concourse URL using the same username and password as your Concourse admin user. We put in a default dashboard that tracks

- Build times
- CPU usage
- Containers
- Disk usage
