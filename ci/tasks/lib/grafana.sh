#!/usr/bin/env bash

# Check that grafana is actually there
function assertGrafanaPresent() {
    curl -ksLo/dev/null --fail https://"$domain":3000
    echo "Grafana is running on Port 3000"
}

# Check that the expected dashboard is present
function assertConcourseDashboardPresent() {
    uid=$(curl -ks --fail "https://${username}:${password}@${domain}:3000/api/search?query=concourse" | jq -r '.[0].uid')
    dashboardType=$(curl -ks --fail "https://${username}:${password}@${domain}:3000/api/dashboards/uid/${uid}" | jq -r '.meta.type')
    dashboardLength=$(curl -ks --fail "https://${username}:${password}@${domain}:3000/api/dashboards/uid/${uid}" | jq -r '.dashboard | length')
    [[ "${dashboardType}" = "db" ]]
    [[ ${dashboardLength} -gt 0 ]]

    echo "Grafana dashboard is present"
}
