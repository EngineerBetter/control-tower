#!/bin/bash
# shellcheck disable=SC1091

# Disabling SC1091 because shellcheck can't find our sourced files

source control-tower/ci/tasks/lib/set-flags.sh
source control-tower/ci/tasks/lib/assert-iaas.sh
source control-tower/ci/tasks/lib/verbose.sh
source control-tower/ci/tasks/lib/id.sh
source control-tower/ci/tasks/lib/pipeline.sh
source control-tower/ci/tasks/lib/trap.sh
source control-tower/ci/tasks/lib/credhub.sh
source control-tower/ci/tasks/lib/grafana.sh
source control-tower/ci/tasks/lib/domain.sh
