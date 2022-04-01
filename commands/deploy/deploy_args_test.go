package deploy_test

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/EngineerBetter/control-tower/commands/deploy"
)

func TestDeployArgs_Validate(t *testing.T) {
	defaultFields := Args{
		AllowIPs:                  "0.0.0.0",
		Region:                    "eu-west-1",
		DBSize:                    "small",
		DBSizeIsSet:               false,
		Domain:                    "",
		BitbucketAuthClientID:     "",
		BitbucketAuthClientSecret: "",
		GithubAuthClientID:        "",
		GithubAuthClientSecret:    "",
		MicrosoftAuthClientID:     "",
		MicrosoftAuthClientSecret: "",
		MicrosoftAuthTenant:       "",
		NoMetricsIsSet:            false,
		IAAS:                      "AWS",
		IAASIsSet:                 true,
		SelfUpdate:                false,
		TLSCert:                   "",
		TLSKey:                    "",
		WebSize:                   "small",
		WorkerCount:               1,
		WorkerSize:                "xlarge",
		WorkerType:                "",
		WorkerTypeIsSet:           false,
	}
	tests := []struct {
		name         string
		modification func() Args
		outcomeCheck func(Args) bool
		wantErr      bool
		expectedErr  string
	}{
		{
			name: "Default args",
			modification: func() Args {
				return defaultFields
			},
			wantErr: false,
		},
		{
			name: "All cert fields should be set",
			modification: func() Args {
				args := defaultFields
				args.TLSCert = "a cool cert"
				args.TLSKey = "a cool key"
				args.Domain = "a cool domain"
				return args
			},
			wantErr: false,
		},
		{
			name: "TLSCert cannot be set without TLSKey",
			modification: func() Args {
				args := defaultFields
				args.TLSCert = "a cool cert"
				args.Domain = "a cool domain"
				return args
			},
			wantErr:     true,
			expectedErr: "--tls-cert requires --tls-key to also be provided",
		},
		{
			name: "IAAS not set",
			modification: func() Args {
				args := defaultFields
				args.IAASIsSet = false
				return args
			},
			wantErr:     true,
			expectedErr: "--iaas flag not set",
		},
		{
			name: "TLSKey cannot be set without TLSCert",
			modification: func() Args {
				args := defaultFields
				args.TLSKey = "a cool key"
				args.Domain = "a cool domain"
				return args
			},
			wantErr:     true,
			expectedErr: "--tls-key requires --tls-cert to also be provided",
		},
		{
			name: "TLSKey and TLSCert require a domain",
			modification: func() Args {
				args := defaultFields
				args.TLSKey = "a cool key"
				args.TLSCert = "a cool cert"
				return args
			},
			wantErr:     true,
			expectedErr: "custom certificates require --domain to be provided",
		},
		{
			name: "Worker count must be positive",
			modification: func() Args {
				args := defaultFields
				args.WorkerCount = 0
				return args
			},
			wantErr:     true,
			expectedErr: "minimum number of workers is 1",
		},
		{
			name: "Worker size must be a known value",
			modification: func() Args {
				args := defaultFields
				args.WorkerSize = "bananas"
				return args
			},
			wantErr:     true,
			expectedErr: fmt.Sprintf("unknown worker size: `bananas`. Valid sizes are: %v", WorkerSizes),
		},
		{
			name: "Web size must be a known value",
			modification: func() Args {
				args := defaultFields
				args.WebSize = "bananas"
				return args
			},
			wantErr:     true,
			expectedErr: fmt.Sprintf("unknown web node size: `bananas`. Valid sizes are: %v", WebSizes),
		},
		{
			name: "DB size must be a known value",
			modification: func() Args {
				args := defaultFields
				args.DBSize = "bananas"
				return args
			},
			wantErr:     true,
			expectedErr: fmt.Sprintf("unknown DB size: `bananas`. Valid sizes are:"),
		},
		{
			name: "Github ID requires Github Secret",
			modification: func() Args {
				args := defaultFields
				args.GithubAuthClientID = "an id"
				return args
			},
			wantErr:     true,
			expectedErr: "--github-auth-client-id requires --github-auth-client-secret to also be provided",
		},
		{
			name: "Github Secret requires Github ID",
			modification: func() Args {
				args := defaultFields
				args.GithubAuthClientSecret = "super secret"
				return args
			},
			wantErr:     true,
			expectedErr: "--github-auth-client-secret requires --github-auth-client-id to also be provided",
		},
		{
			name: "Tags should be in the format 'key=value'",
			modification: func() Args {
				args := defaultFields
				args.Tags = []string{"Key=Value", "Cheese=Ham"}
				return args
			},
			wantErr: false,
		},
		{
			name: "Invalid tags should throw a helpful error",
			modification: func() Args {
				args := defaultFields
				args.Tags = []string{"not a real tag"}
				return args
			},
			wantErr:     true,
			expectedErr: "`not a real tag` is not in the format `key=value`",
		},
		{
			name: "Both public-subnet-range and private-subnet-range are required when either is provided",
			modification: func() Args {
				args := defaultFields
				args.PrivateCIDR = "10.0.1.0/24"
				return args
			},
			wantErr:     true,
			expectedErr: "both --public-subnet-range and --private-subnet-range are required when either is provided",
		},
		{
			name: "Valid worker-type should succeed",
			modification: func() Args {
				args := defaultFields
				args.WorkerTypeIsSet = true
				args.WorkerType = "m5a"
				return args
			},
			wantErr: false,
		},
		{
			name: "Invalid worker-type should throw a helpful error",
			modification: func() Args {
				args := defaultFields
				args.WorkerTypeIsSet = true
				args.WorkerType = "m5b"
				return args
			},
			wantErr:     true,
			expectedErr: "worker-type m5b is invalid: must be one of m4, m5, or m5a",
		},
		{
			name: "Setting worker-type and and iaas other than AWS should throw a helpful error",
			modification: func() Args {
				args := defaultFields
				args.WorkerTypeIsSet = true
				args.WorkerType = "m5"
				args.IAAS = "GCP"
				return args
			},
			wantErr:     true,
			expectedErr: "worker-type is only defined on AWS",
		},
		{
			name: "Setting --no-metrics and --influxdb-retention-period together is an error",
			modification: func() Args {
				args := defaultFields
				args.NoMetricsIsSet = true
				args.InfluxDbRetentionIsSet = true
				args.InfluxDbRetention = "3d"
				return args
			},
			wantErr:     true,
			expectedErr: "no-metrics is invalid when used with influxdb-retention-period",
		},
		{
			name: "-invalid is not a valid GitHub user for main auth",
			modification: func() Args {
				args := defaultFields
				args.GithubAuthIsSet = true
				args.MainGithubAuthIsSet = true
				args.MainGithubUsers = "-invalid"
				args.MainGithubUsersIsSet = true
				return args
			},
			wantErr:     true,
			expectedErr: "Invalid user \"-invalid\" provided to --main-team-github-users",
		},
		{
			name: "valid users not passed as comma separated are not valid for main auth",
			modification: func() Args {
				args := defaultFields
				args.GithubAuthIsSet = true
				args.MainGithubAuthIsSet = true
				args.MainGithubUsers = "a-user other-user"
				args.MainGithubUsersIsSet = true
				return args
			},
			wantErr:     true,
			expectedErr: "Invalid user \"a-user other-user\" provided to --main-team-github-users",
		},
		{
			name: "valid users passed as comma serparated with spaces are valid for main auth",
			modification: func() Args {
				args := defaultFields
				args.GithubAuthIsSet = true
				args.MainGithubAuthIsSet = true
				args.MainGithubUsers = "a-user, other-user"
				args.MainGithubUsersIsSet = true
				return args
			},
			wantErr: false,
		},
		{
			name: "not-an-@rg is not a valid GitHub org for main auth",
			modification: func() Args {
				args := defaultFields
				args.GithubAuthIsSet = true
				args.MainGithubAuthIsSet = true
				args.MainGithubOrgs = "not-an-@rg"
				args.MainGithubOrgsIsSet = true
				return args
			},
			wantErr:     true,
			expectedErr: "Invalid org \"not-an-@rg\" provided to --main-team-github-orgs",
		},
		{
			name: "teams without orgs are not a valid for main auth",
			modification: func() Args {
				args := defaultFields
				args.GithubAuthIsSet = true
				args.MainGithubAuthIsSet = true
				args.MainGithubTeams = "a-team"
				args.MainGithubTeamsIsSet = true
				return args
			},
			wantErr:     true,
			expectedErr: "Invalid team \"a-team\" does not contain org",
		},
		{
			name: "not-a-tea^ is not a valid GitHub team for main auth",
			modification: func() Args {
				args := defaultFields
				args.GithubAuthIsSet = true
				args.MainGithubAuthIsSet = true
				args.MainGithubTeams = "valid-org:not-a-te@m"
				args.MainGithubTeamsIsSet = true
				return args
			},
			wantErr:     true,
			expectedErr: "Invalid team \"valid-org:not-a-te@m\" provided to --main-team-github-teams",
		},
		{
			name: "invalid-*rg is not a valid GitHub org for a GitHub team in main auth",
			modification: func() Args {
				args := defaultFields
				args.GithubAuthIsSet = true
				args.MainGithubAuthIsSet = true
				args.MainGithubTeams = "valid-org:valid-team, invalid-*rg:other-team"
				args.MainGithubTeamsIsSet = true
				return args
			},
			wantErr:     true,
			expectedErr: "Invalid org \"invalid-*rg\" provided for team \"other-team\" in --main-team-github-teams",
		},
		{
			name: "comma separated valid org:team combos without spaces are valid for main auth",
			modification: func() Args {
				args := defaultFields
				args.GithubAuthIsSet = true
				args.MainGithubAuthIsSet = true
				args.MainGithubTeams = "valid-org:valid-team,other-org:other-team"
				args.MainGithubTeamsIsSet = true
				return args
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.modification()
			err := args.Validate()
			if (err != nil) != tt.wantErr || (err != nil && tt.wantErr && !strings.Contains(err.Error(), tt.expectedErr)) {
				if err != nil {
					t.Errorf("DeployArgs.Validate() %v test failed.\nFailed with error = %v,\nExpected error = %v,\nShould fail %v\nWith args: %#v", tt.name, err.Error(), tt.expectedErr, tt.wantErr, args)
				} else {
					t.Errorf("DeployArgs.Validate() %v test failed.\nShould fail %v\nWith args: %#v", tt.name, tt.wantErr, args)
				}
			}
			if tt.outcomeCheck != nil {
				if tt.outcomeCheck(args) {
					t.Errorf("DeployArgs.Validate() %v test failed.\nShould fail %v\nWith args: %#v", tt.name, tt.wantErr, args)
				}
			}
		})
	}
}

func TestDeployArgs_MarkSetFlags(t *testing.T) {
	tests := []struct {
		name                    string
		specifiedFlags          []string
		wantErr                 bool
		expectedGithubAuthIsSet bool
	}{
		{
			name:                    "GithubAuthIsSet is true when both ClientID and ClientSecret are set",
			specifiedFlags:          []string{"github-auth-client-id", "github-auth-client-secret"},
			wantErr:                 false,
			expectedGithubAuthIsSet: true,
		},
		{
			name:                    "GithubAuthIsSet is false when ClientID is set and ClientSecret is not",
			specifiedFlags:          []string{"github-auth-client-id"},
			wantErr:                 false,
			expectedGithubAuthIsSet: false,
		},
		{
			name:                    "GithubAuthIsSet is false when ClientID is not set and ClientSecret is",
			specifiedFlags:          []string{"github-auth-client-secret"},
			wantErr:                 false,
			expectedGithubAuthIsSet: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Args{}
			c := NewFakeFlagSetChecker([]string{"github-auth-client-id", "github-auth-client-secret"}, tt.specifiedFlags)
			if err := a.MarkSetFlags(&c); (err != nil) != tt.wantErr {
				t.Errorf("DeployArgs.MarkSetFlags() error = %v, wantErr %v", err, tt.wantErr)
			}

			if a.GithubAuthIsSet != tt.expectedGithubAuthIsSet {
				t.Errorf("DeployArgs.MarkSetFlags() set GitHubAuthIsSet to %v, was expecting %v", a.GithubAuthIsSet, tt.expectedGithubAuthIsSet)
			}
		})
	}
}

type FakeFlagSetChecker struct {
	names          []string
	specifiedFlags []string
}

func NewFakeFlagSetChecker(names, specifiedFlags []string) FakeFlagSetChecker {
	return FakeFlagSetChecker{
		names:          names,
		specifiedFlags: specifiedFlags,
	}
}

func (f *FakeFlagSetChecker) IsSet(desired string) bool {
	for _, flag := range f.specifiedFlags {
		if desired == flag {
			return true
		}
	}
	return false
}

func (f *FakeFlagSetChecker) FlagNames() (names []string) {
	return names
}
