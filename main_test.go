package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFormat(t *testing.T) {
	testCases := []struct {
		Name    string
		Message string
		want    Format
		wantErr error
	}{
		{
			Name:    "happy path",
			Message: "feat(test): samples",
			want: Format{
				Type:    "feat",
				Scope:   "test",
				Subject: "samples",
			},
			wantErr: nil,
		},
		{
			Name:    "happy path 2",
			Message: "feat(test):                         samples",
			want: Format{
				Type:    "feat",
				Scope:   "test",
				Subject: "samples",
			},
			wantErr: nil,
		},
		{
			Name:    "invalid format",
			Message: "feat(test):samples",
			wantErr: ErrFormat,
		},
		{
			Name:    "scope empty",
			Message: "feat: global",
			want: Format{
				Type:    "feat",
				Scope:   "",
				Subject: "global",
			},
			wantErr: nil,
		},
		{
			Name:    "scope empty 2",
			Message: "feat(): global",
			wantErr: ErrScope,
		},
		{
			Name:    "type empty",
			Message: "(test): test",
			wantErr: ErrFormat,
		},
		{
			Name:    "subject empty 1",
			Message: "feat(test):",
			wantErr: ErrFormat,
		},
		{
			Name:    "subject empty 2",
			Message: "feat(test):   ",
			wantErr: ErrFormat,
		},
		{
			Name: "subject empty 3",
			Message: "feat(test):        		 ",
			wantErr: ErrFormat,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			f, err := NewFormat(tc.Message)
			if tc.wantErr == nil {
				assert.NoError(t, err)
			}
			if err != nil {
				assert.Equal(t, tc.wantErr.Error(), err.Error())
				return
			}

			assert.Equal(t, tc.want, f)
		})
	}
}

func TestVerify(t *testing.T) {
	testCases := []struct {
		Name    string
		Message string
		want    Format
		wantErr error
	}{
		{
			Name:    "happy path",
			Message: "feat(test): samples",
			want: Format{
				Type:    "feat",
				Scope:   "test",
				Subject: "samples",
			},
			wantErr: nil,
		},
		{
			Name:    "invalid type",
			Message: "invalid(test): test",
			wantErr: ErrType,
		},
		{
			Name:    "invalid style",
			Message: "feat(Hest): test",
			wantErr: ErrStyle,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			f, err := NewFormat(tc.Message)
			if err != nil {
				assert.NoError(t, err)
			}
			c, _ := NewConfig("")

			err = f.Verify(c)
			if tc.wantErr == nil {
				assert.NoError(t, err)
			}
			if err != nil {
				assert.Equal(t, tc.wantErr.Error(), err.Error())
				return
			}

			assert.Equal(t, tc.want, f)
		})
	}
}
