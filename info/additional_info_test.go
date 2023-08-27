package info

import "testing"

type test struct {
	Test           string
	AdditionalInfo AdditionalInfo
}

func TestAdditionalInfo_Set(t *testing.T) {
	info := &AdditionalInfo{}
	type args struct {
		key   string
		value interface{}
	}
	tests := []struct {
		name string
		data *AdditionalInfo
		args args
	}{
		{
			name: "ok",
			data: info,
			args: args{
				key:   "test",
				value: "value",
			},
		},
		{
			name: "ok",
			data: info,
			args: args{
				key:   "test2",
				value: "value2",
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.data.Set(tt.args.key, tt.args.value)
			t.Log(tt.data)
			t.Log(info.GetString(tt.args.key))
		})
	}
}

func TestStructAdditonalInfo(t *testing.T) {
	var info test
	info.AdditionalInfo.Set("key", "value")
	t.Log(info.AdditionalInfo)
	t.Log(info.AdditionalInfo.GetString("key"))
}
