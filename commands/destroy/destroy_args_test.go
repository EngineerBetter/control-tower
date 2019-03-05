package destroy_test

import (
	"strings"
	"testing"

	. "github.com/EngineerBetter/control-tower/commands/destroy"
)

func TestDestroyArgs_Validate(t *testing.T) {
	defaultFields := Args{
		Region:    "eu-west-1",
		IAAS:      "AWS",
		IAASIsSet: true,
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
			name: "IAAS not set",
			modification: func() Args {
				args := defaultFields
				args.IAASIsSet = false
				return args
			},
			wantErr:     true,
			expectedErr: "--iaas flag not set",
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
