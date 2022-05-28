package generator

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_hashToBase62Generator_GenerateIdFromString(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "str to id",
			args:    args{str: "https://yandex.ru"},
			want:    "9vnMM4Hf4Os",
			wantErr: false,
		},
		{
			name: "long str to id",
			args: args{str: "some very very very very very very very very very" +
				" very very very long very very very very very very very very very" +
				" very very very long very very very very very very very very very " +
				"very very very long very very very very very very very very very very " +
				"very very long very very very very very very very very very very very very" +
				" long very very very very very very very very very very very very long very " +
				"very very very very very very very very very very very long very very very very" +
				" very very very very very very very very long string"},
			want:    "bwA8a96q4kE",
			wantErr: false,
		},
		{
			name:    "empty string",
			args:    args{str: ""},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ha := HashGenerator{}
			got, err := ha.GenerateIDFromString(tt.args.str)
			if !tt.wantErr {
				require.NoError(t, err)
			}
			assert.Equal(t, got, tt.want)
		})
	}
}
