package websockets

import "testing"

func Test_getAcceptKey(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "known result from mozilla guide",
			args: args{key: "dGhlIHNhbXBsZSBub25jZQ=="},
			want: "s3pPLMBiTxaQ9kYGzzhZRbK+xOo=",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getAcceptKey(tt.args.key); got != tt.want {
				t.Errorf("getAcceptKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
