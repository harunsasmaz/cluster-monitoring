package cluster

import "testing"

func TestLabelParams_AsFilter(t *testing.T) {
	type fields struct {
		Label string
		Value string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "should_match",
			fields: fields{
				Label: "label",
				Value: "value",
			},
			want: "label=value",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &LabelParams{
				Label: tt.fields.Label,
				Value: tt.fields.Value,
			}
			if got := p.AsFilter(); got != tt.want {
				t.Errorf("AsFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToLabelSelector(t *testing.T) {
	type args struct {
		params []LabelParams
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "should_match",
			args: args{
				params: []LabelParams{
					{
						Label: "label",
						Value: "value",
					},
					{
						Label: "label2",
						Value: "value2",
					},
				},
			},
			want: "label=value,label2=value2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toLabelSelector(tt.args.params...); got != tt.want {
				t.Errorf("ToLabelSelector() = %v, want %v", got, tt.want)
			}
		})
	}
}
