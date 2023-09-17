package env

import (
	"reflect"
	"testing"

	awssts "github.com/aws/aws-sdk-go/service/sts"
)

func TestNewEnvironmentWithCredentials(t *testing.T) {
	type args struct {
		env                                        []string
		AccessKeyId, SecretAccessKey, SessionToken string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "simple", args: args{[]string{"A=B", "C=D"}, "key", "secret", "token"},
			want: []string{"A=B", "C=D", "AWS_ACCESS_KEY_ID=key", "AWS_SECRET_ACCESS_KEY=secret", "AWS_SESSION_TOKEN=token"},
		},
		{
			name: "simple", args: args{[]string{"AWS_ACCESS_KEY_ID=B", "C=D"}, "key", "secret", "token"},
			want: []string{"C=D", "AWS_ACCESS_KEY_ID=key", "AWS_SECRET_ACCESS_KEY=secret", "AWS_SESSION_TOKEN=token"},
		},
		{
			name: "override all", args: args{[]string{"AWS_ACCESS_KEY_ID=key", "AWS_SECRET_ACCESS_KEY=secret", "AWS_SESSION_TOKEN=token"}, "new_key", "new_secret", "new_token"},
			want: []string{"AWS_ACCESS_KEY_ID=new_key", "AWS_SECRET_ACCESS_KEY=new_secret", "AWS_SESSION_TOKEN=new_token"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			credentials := awssts.Credentials{
				AccessKeyId:     &tt.args.AccessKeyId,
				SecretAccessKey: &tt.args.SecretAccessKey,
				SessionToken:    &tt.args.SessionToken}

			if got := NewEnvironmentWithCredentials(tt.args.env, &credentials); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewEnvironmentWithCredentials() = %v, want %v", got, tt.want)
			}
		})
	}
}
