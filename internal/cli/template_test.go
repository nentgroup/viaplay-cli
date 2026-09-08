package cli

import "testing"

const testTemplatePath = "/path/to/template"

func TestResolveTestTemplateSource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		args             []string
		templatePathFlag string
		want             string
		wantErr          bool
	}{
		{
			name: "positional source",
			args: []string{"github.com/owner/repo"},
			want: "github.com/owner/repo",
		},
		{
			name:             "deprecated flag only",
			templatePathFlag: testTemplatePath,
			want:             testTemplatePath,
		},
		{
			name:             "positional and matching flag",
			args:             []string{testTemplatePath},
			templatePathFlag: testTemplatePath,
			want:             testTemplatePath,
		},
		{
			name:             "positional and conflicting flag errors",
			args:             []string{"/path/a"},
			templatePathFlag: "/path/b",
			wantErr:          true,
		},
		{
			name:    "neither given errors",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := resolveTestTemplateSource(tt.args, tt.templatePathFlag)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("resolveTestTemplateSource(%v, %q) expected error, got none", tt.args, tt.templatePathFlag)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveTestTemplateSource(%v, %q) unexpected error: %v", tt.args, tt.templatePathFlag, err)
			}
			if got != tt.want {
				t.Errorf("resolveTestTemplateSource(%v, %q) = %q, want %q", tt.args, tt.templatePathFlag, got, tt.want)
			}
		})
	}
}
