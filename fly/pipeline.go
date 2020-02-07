package fly

// Pipeline is interface for self update pipeline
type Pipeline interface {
	BuildPipelineParams(deployment, namespace, region, domain, iaas string) (Pipeline, error)
	GetConfigTemplate() string
}

type PipelineTemplateParams struct {
	ControlTowerVersion string
	Deployment          string
	Domain              string
	Namespace           string
	Region              string
	IaaS                string
}

const selfUpdateResources = `
resources:
- name: control-tower-release
  type: github-release
  source:
    user: engineerbetter
    repository: control-tower
    pre_release: true
`
