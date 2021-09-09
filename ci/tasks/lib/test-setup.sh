#!/bin/bash
# shellcheck disable=SC1091

set -euo pipefail

source control-tower/ci/tasks/lib/assert-iaas.sh
source control-tower/ci/tasks/lib/verbose.sh
source control-tower/ci/tasks/lib/id.sh
source control-tower/ci/tasks/lib/pipeline.sh
source control-tower/ci/tasks/lib/trap.sh
source control-tower/ci/tasks/lib/credhub.sh
source control-tower/ci/tasks/lib/grafana.sh
source control-tower/ci/tasks/lib/domain.sh
source control-tower/ci/tasks/lib/update-fly.sh
